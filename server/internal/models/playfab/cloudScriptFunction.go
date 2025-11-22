package playfab

type CloudScriptFunction[P any, R any] interface {
	Run(P) R
	Name() string
}
