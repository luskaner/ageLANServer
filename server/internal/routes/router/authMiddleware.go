package router

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/luskaner/ageLANServer/common"
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/routes/game/login"
)

const authCacheValidity = 30 * 24 * time.Hour

func AuthMiddlewareOffline(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := r.Context().Value("user").(models.User)
		now := r.Context().Value("time").(time.Time)
		var valid bool
		_ = u.GetAuth().WithReadOnly(func(deadline *time.Time) error {
			valid = !(*deadline).IsZero() && (*deadline).After(now)
			return nil
		})
		if valid {
			next.ServeHTTP(w, r)
		} else {
			login.PlatformLoginError(now, w)
		}
	})
}

func AuthMiddleware(next http.HandlerFunc, gameId string, cached bool) http.Handler {
	hosts := common.GameHostsDirect(gameId)
	hostToIp := make(map[string]string)
	for _, host := range hosts {
		if ip, err := common.DirectHostToIP(host); err == nil {
			hostToIp[host] = ip
		}
	}
	caCert, err := os.ReadFile(filepath.Join(models.EtcFolder, "cacert.pem"))
	if err != nil {
		panic(err)
	}
	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
		panic("Cannot parse internal CA store")
	}
	tlsConfig := &tls.Config{
		RootCAs: caCertPool,
	}
	dialer := &net.Dialer{
		Timeout: 1 * time.Second,
	}
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, port, dialErr := net.SplitHostPort(addr)
			if dialErr != nil {
				return dialer.DialContext(ctx, network, addr)
			}
			if ip, ok := hostToIp[host]; ok {
				addr = net.JoinHostPort(ip, port)
			} else {
				return nil, fmt.Errorf("host %s not resolved", host)
			}
			return dialer.DialContext(ctx, network, addr)
		},
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := r.Context().Value("time").(time.Time)
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			login.PlatformLoginError(t, w)
			return
		}
		if len(bodyBytes) == 0 && r.PostForm != nil {
			bodyBytes = []byte(r.PostForm.Encode())
		}
		errorHandler := func() {
			if cached {
				AuthMiddlewareOffline(next).ServeHTTP(w, r)
			} else {
				login.PlatformLoginError(t, w)
			}
		}
		doRequest := func(c *http.Client, r *http.Request) (respAny i.A, ok bool) {
			resp, localErr := c.Do(r)
			if localErr != nil {
				errorHandler()
				return
			}
			if resp.StatusCode != http.StatusOK {
				errorHandler()
				return
			}
			respBody, localErr := io.ReadAll(resp.Body)
			if localErr != nil {
				errorHandler()
				return
			}
			_ = resp.Body.Close()
			localErr = json.Unmarshal(respBody, &respAny)
			if localErr != nil || len(respAny) == 0 {
				errorHandler()
				return
			}

			errorCode, ok := respAny[0].(float64)
			if !ok {
				errorHandler()
				return
			}

			if int(errorCode) != 0 {
				i.RawJSON(&w, respBody)
				return
			}

			ok = true
			return
		}
		authReq := r.Clone(context.Background())
		authReq.RequestURI = ""
		authReq.RemoteAddr = ""
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		authReq.Header.Set("User-Agent", common.UserAgent())
		authReq.URL.Scheme = "https"
		authReq.URL.Host = r.Host
		authReq.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		authReq.ContentLength = int64(len(bodyBytes))
		authReq.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewBuffer(bodyBytes)), nil
		}
		authReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		authReq.Form = nil
		jar, err := cookiejar.New(nil)
		if err != nil {
			panic(err)
		}
		client := &http.Client{
			Timeout:   1 * time.Second,
			Jar:       jar,
			Transport: transport,
		}
		authResp, ok := doRequest(client, authReq)
		callNum := 2
		if !ok {
			return
		}
		upstreamSessionId := authResp[1].(string)
		defer func() {
			data := url.Values{}
			data.Add("callNum", strconv.Itoa(callNum))
			data.Add("connect_id", upstreamSessionId)
			data.Add("sessionID", upstreamSessionId)
			encodedData := data.Encode()
			dataReader := strings.NewReader(encodedData)
			logoutReq := r.Clone(context.Background())
			logoutReq.RequestURI = ""
			logoutReq.RemoteAddr = ""
			logoutReq.Header.Set("User-Agent", common.UserAgent())
			logoutReq.URL.Scheme = "https"
			logoutReq.URL.Host = r.Host
			logoutReq.URL.RawQuery = ""
			logoutReq.URL.Path = "/game/login/logout"
			logoutReq.Body = io.NopCloser(dataReader)
			logoutReq.ContentLength = int64(dataReader.Len())
			logoutReq.GetBody = func() (io.ReadCloser, error) {
				reader := strings.NewReader(encodedData)
				return io.NopCloser(reader), nil
			}
			logoutReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			logoutReq.Form = nil
			_, _ = doRequest(client, logoutReq)
		}()
		u := r.Context().Value("user").(models.User)
		if cached {
			_ = u.GetAuth().WithReadWrite(func(deadline *time.Time) error {
				*deadline = t.Add(authCacheValidity)
				return nil
			})
		}
		req, err := http.NewRequest(
			"GET",
			fmt.Sprintf(
				"https://%s/%s",
				r.Host,
				"game/item/getInventoryByProfileIDs",
			),
			nil,
		)
		if err != nil {
			errorHandler()
			return
		}
		req.Header.Set("User-Agent", common.UserAgent())
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Pragma", "no-cache")
		q := req.URL.Query()
		q.Add("callNum", strconv.Itoa(callNum))
		q.Add("connect_id", upstreamSessionId)
		q.Add("sessionID", upstreamSessionId)
		q.Add("profileIDs", fmt.Sprintf("[%d]", int32(authResp[5].(i.A)[0].(i.A)[1].(float64))))
		req.URL.RawQuery = q.Encode()

		resp, ok := doRequest(client, req)
		if !ok {
			return
		}
		callNum++
		itemsRaw := resp[1].(i.A)[0].(i.A)[1].(i.A)
		_ = u.GetItems().WithReadWrite(func(items *map[int32]*models.MainItem) error {
			clear(*items)
			for _, itemRaw := range itemsRaw {
				item := models.NewMainItemFromRaw(itemRaw.(i.A))
				(*items)[item.GetId()] = item
			}
			return nil
		})
		next.ServeHTTP(w, r)
	})
}
