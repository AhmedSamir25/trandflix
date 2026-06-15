let communities = [];
let stats = null;
let total = 0;
let page = 1;
const perPage = 20;

let searchQuery = "";
let statusFilter = "";
let typeFilter = "";

let searchTimer = null;

const CATEGORY_ICON = {
  movies: "🎬",
  series: "📺",
  books: "📚",
  games: "🎮",
  mixed: "✨",
};

const STATUS_META = {
  approved: { key: "admin.commStatusApproved", fallback: "Active", cls: "comm-badge comm-badge--approved" },
  pending: { key: "admin.commStatusPending", fallback: "Pending", cls: "comm-badge comm-badge--pending" },
  rejected: { key: "admin.commStatusRejected", fallback: "Blocked", cls: "comm-badge comm-badge--blocked" },
};

function categoryLabel(type) {
  const labels = {
    movies: t("communities.catMovies"),
    series: t("communities.catSeries"),
    books: t("communities.catBooks"),
    games: t("communities.catGames"),
    mixed: t("communities.catMixed"),
  };
  return labels[type] || `${CATEGORY_ICON[type] || ""} ${type || ""}`.trim();
}

function statusLabel(status) {
  const meta = STATUS_META[status] || { key: "", fallback: status, cls: "comm-badge" };
  return { cls: meta.cls, label: meta.key ? t(meta.key) : meta.fallback };
}

function formatDate(value) {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  return date.toLocaleDateString();
}

function renderStats() {
  document.getElementById("statTotal").textContent = Number(stats?.total || 0).toLocaleString();

  const byStatus = {};
  (stats?.by_status || []).forEach((entry) => {
    byStatus[entry.status] = Number(entry.count || 0);
  });

  document.getElementById("statApproved").textContent = Number(byStatus.approved || 0).toLocaleString();
  document.getElementById("statPending").textContent = Number(byStatus.pending || 0).toLocaleString();
  document.getElementById("statBlocked").textContent = Number(byStatus.rejected || 0).toLocaleString();

  const membersBadge = document.getElementById("membersSumBadge");
  const membersSum = Number(stats?.members_sum || 0);
  membersBadge.hidden = false;
  membersBadge.textContent = `👥 ${membersSum.toLocaleString()}`;

  renderTypeBars();
}

function renderTypeBars() {
  const container = document.getElementById("typeBars");
  if (!container) return;

  const counts = {};
  (stats?.by_type || []).forEach((entry) => {
    counts[entry.type] = Number(entry.count || 0);
  });

  const totalAll = Object.values(counts).reduce((sum, value) => sum + value, 0) || 1;
  const types = ["movies", "series", "books", "games", "mixed"];

  container.innerHTML = types
    .map((type) => {
      const count = counts[type] || 0;
      const percent = Math.round((count / totalAll) * 100);
      return `
        <div class="type-bar-row">
          <span class="type-bar-label">${escapeHtml(categoryLabel(type))}<em>${count.toLocaleString()}</em></span>
          <div class="type-bar-track">
            <span style="width:${percent}%"></span>
          </div>
        </div>`;
    })
    .join("");
}

function renderTable() {
  const body = document.getElementById("communitiesTableBody");
  const empty = document.getElementById("communitiesEmpty");
  if (!body) return;

  if (!communities.length) {
    body.innerHTML = "";
    if (empty) empty.hidden = false;
    return;
  }
  if (empty) empty.hidden = true;

  body.innerHTML = communities
    .map((community) => {
      const avatar = community.avatar_image || getFallbackImage(community.name);
      const owner = community.user?.name || community.user?.email || t("communities.unknown");
      const status = statusLabel(community.status);
      const created = formatDate(community.created_at);
      const isBlocked = community.status === "rejected";
      const members = Number(community.members_count || 0).toLocaleString();

      return `
        <tr data-id="${community.id}">
          <td>
            <div class="comm-cell">
              <img class="comm-cell__avatar" src="${escapeHtml(avatar)}" alt="" loading="lazy" />
              <div class="comm-cell__text">
                <strong>${escapeHtml(community.name)}</strong>
                ${community.is_private ? `<span class="comm-private" data-i18n="admin.commPrivate">Private</span>` : ""}
              </div>
            </div>
          </td>
          <td><span class="info-tag">${escapeHtml(categoryLabel(community.category_type))}</span></td>
          <td><span class="comm-owner">${escapeHtml(owner)}</span></td>
          <td><span class="comm-members">👥 ${escapeHtml(members)}</span></td>
          <td><span class="${status.cls}">${escapeHtml(status.label)}</span></td>
          <td><span class="comm-date">${escapeHtml(created)}</span></td>
          <td>
            <div class="comm-actions">
              <a class="text-btn" href="/pages/community.html?slug=${encodeURIComponent(community.slug)}" target="_blank" rel="noopener" data-i18n="admin.viewCommunity">View</a>
              ${isBlocked
                ? `<button class="text-btn" type="button" data-unblock="${community.id}" data-i18n="admin.commUnblock">Unblock</button>`
                : `<button class="text-btn" type="button" data-block="${community.id}" data-i18n="admin.commBlock">Block</button>`}
              <button class="text-btn danger-btn" type="button" data-delete="${community.id}" data-i18n="admin.deleteItem">Delete</button>
            </div>
          </td>
        </tr>`;
    })
    .join("");
}

function renderCountBadge() {
  const badge = document.getElementById("communitiesCountBadge");
  if (!badge) return;
  badge.hidden = false;
  badge.textContent = Number(total || 0).toLocaleString();
}

function renderPagination() {
  const container = document.getElementById("commPagination");
  if (!container) return;

  const pages = Math.max(1, Math.ceil(total / perPage));
  if (pages <= 1) {
    container.innerHTML = "";
    return;
  }

  const parts = [];
  parts.push(
    `<button class="page-btn" type="button" data-page="${page - 1}"${page === 1 ? " disabled" : ""}>‹</button>`,
  );
  for (let p = 1; p <= pages; p++) {
    parts.push(
      `<button class="page-btn${p === page ? " active" : ""}" type="button" data-page="${p}">${p}</button>`,
    );
  }
  parts.push(
    `<button class="page-btn" type="button" data-page="${page + 1}"${page === pages ? " disabled" : ""}>›</button>`,
  );
  container.innerHTML = parts.join("");
}

function render() {
  renderTable();
  renderCountBadge();
  renderPagination();
}

async function loadStats() {
  const data = await fetchJson("/api/admin/communities/stats", { headers: authHeaders() });
  stats = data?.data?.stats || null;
  renderStats();
}

async function loadCommunities() {
  clearNotice("pageError");
  const params = new URLSearchParams();
  params.set("page", String(page));
  params.set("per_page", String(perPage));
  if (searchQuery) params.set("q", searchQuery);
  if (statusFilter) params.set("status", statusFilter);
  if (typeFilter) params.set("category_type", typeFilter);

  const data = await fetchJson(`/api/admin/communities?${params.toString()}`, { headers: authHeaders() });
  communities = Array.isArray(data?.data?.communities) ? data.data.communities : [];
  total = Number(data?.data?.total || 0);
  render();
}

function reloadKeepingPage() {
  loadCommunities().catch(showPageError);
  loadStats().catch(() => {});
}

async function deleteCommunity(id) {
  const community = communities.find((entry) => Number(entry.id) === Number(id));
  const name = community?.name || t("communities.communityFallback");
  if (!window.confirm(t("admin.confirmDeleteCommunity").replace("{name}", name))) return;

  clearNotice("pageError");
  await fetchJson(`/api/admin/communities/${id}`, {
    method: "DELETE",
    headers: authHeaders(),
  });
  setNotice("pageNotice", t("admin.commDeleted").replace("{name}", name), "success");
  reloadKeepingPage();
  window.refreshPendingCommunitiesCount?.();
}

async function changeStatus(id, action, successKey) {
  const community = communities.find((entry) => Number(entry.id) === Number(id));
  const name = community?.name || t("communities.communityFallback");
  clearNotice("pageError");
  await fetchJson(`/api/admin/communities/${id}/${action}`, {
    method: "POST",
    headers: authHeaders(),
  });
  setNotice("pageNotice", t(successKey).replace("{name}", name), "success");
  reloadKeepingPage();
  window.refreshPendingCommunitiesCount?.();
}

function bindFilters() {
  document.getElementById("statusFilters")?.addEventListener("click", (event) => {
    const btn = event.target.closest("[data-status]");
    if (!btn) return;
    statusFilter = btn.dataset.status;
    document.querySelectorAll("#statusFilters .filter-btn").forEach((node) => {
      node.classList.toggle("active", node === btn);
    });
    page = 1;
    loadCommunities().catch(showPageError);
  });

  document.getElementById("typeFilters")?.addEventListener("click", (event) => {
    const btn = event.target.closest("[data-type]");
    if (!btn) return;
    typeFilter = btn.dataset.type;
    document.querySelectorAll("#typeFilters .filter-btn").forEach((node) => {
      node.classList.toggle("active", node === btn);
    });
    page = 1;
    loadCommunities().catch(showPageError);
  });

  const search = document.getElementById("commSearch");
  search?.addEventListener("input", (event) => {
    searchQuery = event.target.value.trim();
    clearTimeout(searchTimer);
    searchTimer = setTimeout(() => {
      page = 1;
      loadCommunities().catch(showPageError);
    }, 300);
  });
}

function bindTableActions() {
  document.getElementById("communitiesTableBody")?.addEventListener("click", (event) => {
    const blockBtn = event.target.closest("[data-block]");
    if (blockBtn) {
      changeStatus(blockBtn.dataset.block, "block", "admin.commBlocked").catch(showPageError);
      return;
    }
    const unblockBtn = event.target.closest("[data-unblock]");
    if (unblockBtn) {
      changeStatus(unblockBtn.dataset.unblock, "unblock", "admin.commUnblocked").catch(showPageError);
      return;
    }
    const deleteBtn = event.target.closest("[data-delete]");
    if (deleteBtn) {
      deleteCommunity(deleteBtn.dataset.delete).catch(showPageError);
    }
  });

  document.getElementById("commPagination")?.addEventListener("click", (event) => {
    const btn = event.target.closest("[data-page]");
    if (!btn || btn.disabled) return;
    const next = Number(btn.dataset.page);
    if (!next || next === page) return;
    page = next;
    loadCommunities().catch(showPageError);
  });
}

window.addEventListener("DOMContentLoaded", async () => {
  if (!requireAdmin()) return;

  bindLogout();
  highlightActiveNav();
  bindFilters();
  bindTableActions();

  document.getElementById("refreshCommunitiesBtn")?.addEventListener("click", () => {
    loadStats().catch(() => {});
    loadCommunities().catch(showPageError);
  });

  window.addEventListener("trendflix:languagechange", () => {
    renderStats();
    render();
  });

  try {
    await Promise.all([loadStats(), loadCommunities()]);
  } catch (error) {
    showPageError(error);
  }
});
