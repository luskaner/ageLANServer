(() => {
  const q = (selector) => document.querySelector(selector);
  const escape = (value) => String(value ?? "").replace(/[&<>"']/g, (char) => ({
    "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;"
  })[char]);
  const admin = Boolean(window.ageServerAdmin);
  let csrf = "";
  let games = [];
  let config = null;
  const status = (message, error = false) => {
    const node = q("#admin-status");
    if (node) {
      node.textContent = message;
      node.classList.toggle("error", error);
    }
  };
  async function request(url, options = {}) {
    const headers = { ...(options.headers || {}) };
    if (options.body) headers["Content-Type"] = "application/json";
    if (csrf && options.method && options.method !== "GET") headers["X-CSRF-Token"] = csrf;
    const response = await fetch(url, { ...options, headers, credentials: "same-origin" });
    const payload = await response.json().catch(() => ({}));
    if (!response.ok) throw new Error(payload.error || `Request failed (${response.status})`);
    return payload;
  }
  const query = (params) => new URLSearchParams(params).toString();
  const statusPill = (value) => `<span class="status-pill" data-status="${escape(value)}"><i></i>${escape(value)}</span>`;
  function battleServerMarkup(server, details = false) {
    const heading = details
      ? `<div class="section-title"><div><p class="eyebrow">${escape(server.game)} · ${escape(server.region)} · ${escape(server.source)}</p><h3>${escape(server.name)}</h3></div>${statusPill(server.status)}</div>`
      : `<div class="section-title"><strong>${escape(server.name)}</strong>${statusPill(server.status)}</div><p>${escape(server.region)} · ${escape(server.source)}</p>`;
    const info = `<dl class="details"><dt>Address</dt><dd><code>${escape(server.ipv4 || "—")}</code></dd><dt>Battle port</dt><dd><code>${escape(server.bsPort)}</code></dd><dt>WebSocket</dt><dd><code>${escape(server.webSocketPort)}</code></dd><dt>Out-of-band</dt><dd><code>${server.outOfBandPort ? escape(server.outOfBandPort) : "Not configured"}</code></dd></dl>`;
    const checked = server.checkedAt ? `<small class="small">Last checked ${escape(new Date(server.checkedAt).toLocaleTimeString())}</small>` : '<small class="small">Initial connection check in progress…</small>';
    return `<article class="battle-server-card">${heading}${details ? info + checked : `<code>${escape(server.ipv4 || "—")} · ${escape(server.bsPort)} / ${escape(server.webSocketPort)}</code>`}</article>`;
  }
  function renderBattleServers(servers, target, details = false) {
    if (!target) return;
    target.innerHTML = (servers || []).map((server) => battleServerMarkup(server, details)).join("") ||
      '<p class="empty">No battle servers are currently loaded.</p>';
  }
  function watchPublicPages() {
    if (admin || !window.EventSource) return;
    const stream = new EventSource("/api/events");
    stream.addEventListener("snapshot", async (event) => {
      let snapshot;
      try { snapshot = JSON.parse(event.data); } catch { return; }
      const selected = new URLSearchParams(window.location.search).get("game");
      const game = (snapshot.games || []).find((item) => item.game === selected) || snapshot.games?.[0];
      if (!game) return;
      renderBattleServers(game.battleServerDetails, document.getElementById("live-battle-servers"), window.location.pathname === "/battle-servers");
      for (const node of document.querySelectorAll("[data-live-count]")) {
        const key = node.dataset.liveCount;
        node.textContent = key === "lobbies" ? game.lobbies.length : key === "players" ? game.players.length : game.queue.length;
      }
      const fragmentByID = {
        "live-lobbies": "lobbies",
        "live-queue": "queue",
        "live-matches-table": "matches",
        "live-leaderboard": "leaderboard",
        "live-players": "players",
        "live-maps": "maps",
      };
      await Promise.all(Object.entries(fragmentByID).map(async ([id, fragment]) => {
        const node = document.getElementById(id);
        if (!node) return;
        const response = await fetch(`/fragments/${fragment}?${query({ game: game.game })}`);
        if (response.ok) node.innerHTML = await response.text();
      }));
      const matches = document.getElementById("live-matches");
      if (matches) {
        matches.replaceChildren();
        if (!game.liveMatches.length) {
          const empty = document.createElement("p");
          empty.className = "empty";
          empty.textContent = "No active automatch games.";
          matches.append(empty);
        }
        for (const match of game.liveMatches) {
          const row = document.createElement("div");
          row.className = "match-card";
          const mapLabel = document.createElement("span");
          mapLabel.className = "icon-label";
          const mapIcon = document.createElement("img");
          mapIcon.className = "map-icon";
          mapIcon.src = match.mapIcon || "/static/images/map-placeholder.svg";
          mapIcon.alt = "";
          mapIcon.onerror = () => { mapIcon.onerror = null; mapIcon.src = "/static/images/map-placeholder.svg"; };
          const map = document.createElement("strong");
          map.textContent = match.displayMapName || match.mapName;
          mapLabel.append(mapIcon, map);
          const state = document.createElement("span");
          state.textContent = match.state;
          const players = document.createElement("span");
          players.textContent = (match.players || []).map((player) => `${player.alias} (${player.ip || "-"})`).join(" vs. ");
          row.append(mapLabel, state, players);
          matches.append(row);
        }
      }
    });
  }
  function populateGameSelect(selector) {
    const select = q(selector);
    if (!select) return;
    select.innerHTML = games.map((game) => `<option value="${escape(game.game)}">${escape(game.game)}</option>`).join("");
  }
  async function loadPool() {
    const game = q("#pool-game").value;
    const data = await request(`/admin/api/pool?${query({ game })}`);
    q("#pool-mode").value = data.mapPool?.mode || "random";
    const allowed = new Set((data.mapPool?.maps || []).map((map) => map.id));
    const weights = new Map((data.mapPool?.maps || []).map((map) => [map.id, map.weight || 1]));
    q("#pool-maps").innerHTML = (data.available || []).map((map) =>
      `<label class="pool-entry"><input type="checkbox" data-map-id="${map.id}" ${!data.poolActive || allowed.has(map.id) ? "checked" : ""}><span class="pool-map-label"><img class="map-icon" src="${escape(map.icon || "/static/images/map-placeholder.svg")}" alt="" onerror="this.onerror=null;this.src='/static/images/map-placeholder.svg'"><span>${escape(map.displayName || map.name)}</span></span><input type="number" min="1" max="1000" data-weight-id="${map.id}" value="${weights.get(map.id) || 1}" aria-label="Weight for ${escape(map.displayName || map.name)}"></label>`
    ).join("") || '<p class="empty">No maps available for this game.</p>';
  }
  async function loadStaticServers() {
    const game = q("#battle-game").value;
    const data = await request(`/admin/api/battle-servers?${query({ game })}`);
    const running = games.find((item) => item.game === game)?.battleServers || [];
    q("#battle-servers").innerHTML = (data.servers || []).map((server) =>
      `<div class="admin-item"><span><strong>${escape(server.region)} · ${escape(server.name)}</strong><br><small>${escape(server.ipv4)} · ${server.bsPort}/${server.webSocketPort}/${server.outOfBandPort || "—"}</small><br>${statusPill(running.find((entry) => entry.region === server.region && entry.source === "Static")?.status || "Restart required")}</span><button type="button" data-delete-region="${escape(server.region)}">Remove</button></div>`
    ).join("") || '<p class="empty">No static battle servers saved.</p>';
    q("#battle-servers").querySelectorAll("[data-delete-region]").forEach((button) => {
      button.addEventListener("click", async () => {
        try {
          await request(`/admin/api/battle-servers/${encodeURIComponent(game)}/${encodeURIComponent(button.dataset.deleteRegion)}`, { method: "DELETE" });
          status("Static battle server removed. Restart the server to apply the change.");
          await loadStaticServers();
        } catch (error) { status(error.message, true); }
      });
    });
  }
  function configInput(field, value) {
    const current = value ?? "";
    if (field.type === "bool") return `<input type="checkbox" data-config-path="${escape(field.path)}" ${current ? "checked" : ""}>`;
    if (field.type === "enum") return `<select data-config-path="${escape(field.path)}">${(field.options || []).map((option) => `<option value="${escape(option)}" ${option === current ? "selected" : ""}>${escape(option)}</option>`).join("")}</select>`;
    if (field.type === "strings") return `<input data-config-path="${escape(field.path)}" value="${escape(Array.isArray(current) ? current.join(", ") : current)}">`;
    return `<input data-config-path="${escape(field.path)}" type="${field.type === "int" ? "number" : "text"}" value="${escape(current)}">`;
  }
  async function loadConfig() {
    config = await request("/admin/api/config");
    const groups = new Map();
    for (const field of config.fields || []) {
      if (!groups.has(field.group)) groups.set(field.group, []);
      groups.get(field.group).push(field);
    }
    q("#config-fields").innerHTML = [...groups].map(([group, fields]) =>
      `<details><summary>${escape(group)}</summary>${fields.map((field) =>
        `<label>${escape(field.label)}${configInput(field, config.values[field.path])}${field.help ? `<small class="small">${escape(field.help)}</small>` : ""}</label>`
      ).join("")}</details>`
    ).join("");
    if (!config.editable) q("#config-save").disabled = true;
  }
  function gameAdminMarkup(game) {
    const lobbies = (game.lobbies || []).map((lobby) =>
      `<div class="admin-item"><span>${lobby.automatch ? "Ranked matchmaking" : "Custom room"} · ${escape(lobby.hostAlias)} · ${escape(lobby.mapName)} · ${escape(lobby.ip)} · ${lobby.playerCount}/${lobby.maxPlayers} · observers ${lobby.observersEnabled ? "allowed" : "off"}</span><button type="button" data-close-lobby="${lobby.id}" data-game="${escape(game.game)}">Close</button></div>`
    ).join("") || '<p class="empty">No active lobbies.</p>';
    const queue = (game.queue || []).map((player) =>
      `<div class="admin-item"><span>${escape(player.alias)} · ${escape(player.ip)} · waiting ${player.waitingSeconds}s</span><button type="button" data-kick-user="${player.userId}" data-game="${escape(game.game)}">Remove</button></div>`
    ).join("") || '<p class="empty">No players in queue.</p>';
    const matches = (game.matches || []).map((match) =>
      `<div class="admin-item"><span>${escape(match.mapName)} · ${(match.players || []).map((player) => `${escape(player.alias)} (${escape(player.ip)})`).join(" vs. ")} · ${escape(match.state)}</span></div>`
    ).join("") || '<p class="empty">No active matches.</p>';
    const recent = (game.recentMatches || []).map((match) =>
      `<div class="admin-item"><span><span class="icon-label"><img class="map-icon" src="${escape(match.mapIcon || "/static/images/map-placeholder.svg")}" alt="" onerror="this.onerror=null;this.src='/static/images/map-placeholder.svg'">${escape(match.displayMapName || match.mapName)}</span> · ${(match.playerInfo || []).map((player) => `<span class="icon-label"><img class="civilization-icon" src="${escape(player.civilizationIcon || "/static/images/civs/unknown.svg")}" alt="" onerror="this.onerror=null;this.src='/static/images/civs/unknown.svg'">${escape(player.alias)} — ${escape(player.civilization)}</span> (${player.won ? "Winner" : "Defeated"})`).join(" vs. ")} · ${escape(match.finishedAt)}</span></div>`
    ).join("") || '<p class="empty">No recent matches.</p>';
    return `<details class="admin-game" open><summary>${escape(game.game)} · ${game.lobbies.length} lobbies · ${game.players.length} online · ${game.queue.length} queued · ${game.matches.length} live</summary>
      <p class="small">Game hosts: ${(game.hosts || []).map(escape).join(", ")}</p><h4>Active lobbies</h4>${lobbies}
      <h4>Automatch queue</h4>${queue}<h4>Live matches</h4>${matches}<h4>Recent matches</h4>${recent}</details>`;
  }
  function renderAdminBattleServers() {
    const servers = games.flatMap((game) => game.battleServers || []);
    const node = q("#admin-active-battleservers");
    if (node) node.innerHTML = servers.map((server) => battleServerMarkup(server, true)).join("") ||
      '<p class="empty">No battle servers are currently loaded.</p>';
  }
  function activateAdminTab(name, updateHash = false) {
    const tabs = [...document.querySelectorAll("[data-admin-tab]")];
    const selected = tabs.find((tab) => tab.dataset.adminTab === name) || tabs[0];
    if (!selected) return;
    for (const tab of tabs) {
      const active = tab === selected;
      tab.setAttribute("aria-selected", String(active));
      tab.tabIndex = active ? 0 : -1;
    }
    for (const panel of document.querySelectorAll("[data-admin-panel]")) {
      panel.hidden = panel.dataset.adminPanel !== selected.dataset.adminTab;
    }
    if (updateHash) window.history.replaceState(null, "", `#${selected.getAttribute("aria-controls")}`);
  }
  function bindAdminTabs() {
    const tabs = [...document.querySelectorAll("[data-admin-tab]")];
    for (const [index, tab] of tabs.entries()) {
      tab.addEventListener("click", () => activateAdminTab(tab.dataset.adminTab, true));
      tab.addEventListener("keydown", (event) => {
        let next = index;
        if (event.key === "ArrowRight") next = (index + 1) % tabs.length;
        else if (event.key === "ArrowLeft") next = (index + tabs.length - 1) % tabs.length;
        else if (event.key === "Home") next = 0;
        else if (event.key === "End") next = tabs.length - 1;
        else return;
        event.preventDefault();
        tabs[next].focus();
        activateAdminTab(tabs[next].dataset.adminTab, true);
      });
    }
    const initial = tabs.find((tab) => tab.getAttribute("aria-controls") === window.location.hash.slice(1));
    activateAdminTab(initial?.dataset.adminTab || tabs[0]?.dataset.adminTab);
    window.addEventListener("hashchange", () => {
      const tab = tabs.find((item) => item.getAttribute("aria-controls") === window.location.hash.slice(1));
      if (tab) activateAdminTab(tab.dataset.adminTab);
    });
  }
  function bindGameActions() {
    document.querySelectorAll("[data-close-lobby]").forEach((button) => {
      button.addEventListener("click", async () => {
        if (!window.confirm("Close this lobby and remove its players from matchmaking?")) return;
        try {
          const game = button.dataset.game;
          await request(`/admin/api/lobbies/${button.dataset.closeLobby}/close?${query({ game })}`, { method: "POST" });
          await loadAdmin();
          status("Lobby closed.");
        } catch (error) { status(error.message, true); }
      });
    });
    document.querySelectorAll("[data-kick-user]").forEach((button) => {
      button.addEventListener("click", async () => {
        try {
          const game = button.dataset.game;
          await request(`/admin/api/queue/${button.dataset.kickUser}/kick?${query({ game })}`, { method: "POST" });
          await loadAdmin();
          status("Player removed from the queue.");
        } catch (error) { status(error.message, true); }
      });
    });
  }
  async function loadAdmin() {
    try {
      const state = await request("/admin/api/state");
      csrf = state.csrf;
      games = state.games || [];
      for (const game of games) {
        game.lobbies ||= [];
        game.players ||= [];
        game.queue ||= [];
        game.matches ||= [];
        game.recentMatches ||= [];
        game.hosts ||= [];
        game.battleServers ||= [];
      }
      populateGameSelect("#pool-game");
      populateGameSelect("#battle-game");
      q("#admin-games").innerHTML = games.map(gameAdminMarkup).join("");
      renderAdminBattleServers();
      bindGameActions();
      await Promise.all([loadPool(), loadStaticServers(), loadConfig()]);
      status("Admin data is ready.");
    } catch (error) { status(error.message, true); }
  }
  if (admin) {
    bindAdminTabs();
    loadAdmin();
    window.setInterval(async () => {
      try {
        const state = await request("/admin/api/state");
        games = state.games || [];
        for (const game of games) game.battleServers ||= [];
        renderAdminBattleServers();
        await loadStaticServers();
      } catch (error) { status(error.message, true); }
    }, 10000);
    q("#pool-game").addEventListener("change", () => loadPool().catch((error) => status(error.message, true)));
    q("#battle-game").addEventListener("change", () => loadStaticServers().catch((error) => status(error.message, true)));
    q("#pool-save").addEventListener("click", async () => {
      try {
        const maps = [...q("#pool-maps").querySelectorAll("[data-map-id]")].filter((input) => input.checked).map((input) => {
          const id = Number(input.dataset.mapId);
          const map = games.find((game) => game.game === q("#pool-game").value)?.availableMaps.find((entry) => entry.id === id);
          return { id, name: map?.name || String(id), weight: Number(q(`[data-weight-id="${id}"]`).value) || 1, enabled: true };
        });
        await request("/admin/api/pool", { method: "PUT", body: JSON.stringify({ game: q("#pool-game").value, mode: q("#pool-mode").value, maps }) });
        status("Map pool saved and active.");
        await loadPool();
      } catch (error) { status(error.message, true); }
    });
    q("#pool-reset").addEventListener("click", async () => {
      try {
        await request("/admin/api/pool/reset", { method: "POST", body: JSON.stringify({ game: q("#pool-game").value }) });
        status("Map pool reset to the built-in map set.");
        await loadPool();
      } catch (error) { status(error.message, true); }
    });
    q("#battle-form").addEventListener("submit", async (event) => {
      event.preventDefault();
      const data = Object.fromEntries(new FormData(event.currentTarget).entries());
      for (const field of ["bsPort", "webSocketPort", "outOfBandPort"]) data[field] = Number(data[field]);
      try {
        await request("/admin/api/battle-servers", { method: "POST", body: JSON.stringify({ game: q("#battle-game").value, ...data }) });
        event.currentTarget.reset();
        status("Static battle server saved. Restart the server to load it.");
        await loadStaticServers();
      } catch (error) { status(error.message, true); }
    });
    q("#config-save").addEventListener("click", async () => {
      const changes = {};
      for (const input of q("#config-fields").querySelectorAll("[data-config-path]")) {
        const field = config.fields.find((item) => item.path === input.dataset.configPath);
        changes[field.path] = field.type === "bool" ? input.checked : field.type === "int" ? Number(input.value) : field.type === "strings" ? input.value.split(",").map((item) => item.trim()).filter(Boolean) : input.value;
      }
      try {
        await request("/admin/api/config", { method: "POST", body: JSON.stringify({ changes }) });
        status("Configuration saved. Restart the server to apply changes.");
      } catch (error) { status(error.message, true); }
    });
    q("#password-form").addEventListener("submit", async (event) => {
      event.preventDefault();
      const data = Object.fromEntries(new FormData(event.currentTarget).entries());
      try {
        await request("/admin/api/password", { method: "POST", body: JSON.stringify(data) });
        event.currentTarget.reset();
        status("Admin password updated.");
      } catch (error) { status(error.message, true); }
    });
  }
  watchPublicPages();
})();
