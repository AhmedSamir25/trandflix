let communities = [];

const CATEGORY_ICON = {
  movies: "🎬",
  series: "📺",
  books: "📚",
  games: "🎮",
  mixed: "✨",
};

function categoryLabel(type) {
  const icon = CATEGORY_ICON[type] || "";
  const labels = {
    movies: t("communities.catMovies"),
    series: t("communities.catSeries"),
    books: t("communities.catBooks"),
    games: t("communities.catGames"),
    mixed: t("communities.catMixed"),
  };
  const label = labels[type] || type || "";
  return icon ? `${icon} ${label}` : label;
}

function formatDate(value) {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  return date.toLocaleDateString();
}

function renderList() {
  const list = document.getElementById("communitiesList");
  if (!list) return;

  if (!communities.length) {
    list.innerHTML = `<p class="notice">${escapeHtml(t("admin.noPendingCommunities"))}</p>`;
    return;
  }

  list.innerHTML = communities
    .map((community) => {
      const cover = community.cover_image || "";
      const avatar = community.avatar_image || getFallbackImage(community.name);
      const creator = community.user?.name || community.user?.email || t("communities.unknown");
      const created = formatDate(community.created_at);

      return `
        <article class="catalog-card admin-community-card" data-id="${community.id}">
          <div class="catalog-card__img-wrap">
            ${cover ? `<img class="admin-community-cover" src="${escapeHtml(cover)}" alt="" loading="lazy" />` : ""}
            <img class="catalog-card__avatar admin-community-avatar" src="${escapeHtml(avatar)}" alt="${escapeHtml(community.name)}" loading="lazy" />
            <span class="catalog-card__type">${escapeHtml(categoryLabel(community.category_type))}</span>
          </div>
          <div class="catalog-card__title">${escapeHtml(community.name)}</div>
          <div class="catalog-card__info">
            <span class="info-tag cats">${escapeHtml(t("communities.createdBy"))}: ${escapeHtml(creator)}</span>
            ${created ? `<span class="info-tag date">${escapeHtml(created)}</span>` : ""}
            <span class="info-tag">👥 ${escapeHtml(t("communities.membersCount").replace("{n}", Number(community.members_count || 0).toLocaleString()))}</span>
          </div>
          ${community.description ? `<p class="admin-community-desc">${escapeHtml(community.description)}</p>` : ""}
          ${community.rules ? `<details class="admin-community-rules"><summary>${escapeHtml(t("communities.dRules"))}</summary><p>${escapeHtml(community.rules)}</p></details>` : ""}
          <div class="catalog-card__actions">
            <button class="text-btn catalog-card__action" type="button" data-approve="${community.id}">
              ${escapeHtml(t("admin.approveCommunity"))}
            </button>
            <button class="text-btn danger-btn catalog-card__action" type="button" data-reject="${community.id}">
              ${escapeHtml(t("admin.rejectCommunity"))}
            </button>
            <a class="text-btn catalog-card__action" href="/pages/community.html?slug=${encodeURIComponent(community.slug)}" target="_blank" rel="noopener">
              ${escapeHtml(t("admin.viewCommunity"))}
            </a>
          </div>
        </article>`;
    })
    .join("");
}

function updatePendingCount() {
  const badge = document.getElementById("pendingCountBadge");
  if (!badge) return;
  const count = communities.length;
  badge.hidden = count === 0;
  badge.textContent = String(count);
}

function render() {
  renderList();
  updatePendingCount();
}

async function loadCommunities() {
  clearNotice("pageError");
  const data = await fetchJson("/api/admin/communities/pending?page=1&per_page=50", { headers: authHeaders() });
  communities = Array.isArray(data?.data?.communities) ? data.data.communities : [];
  render();
}

async function setCommunityStatus(id, action) {
  clearNotice("pageError");
  await fetchJson(`/api/admin/communities/${id}/${action}`, {
    method: "POST",
    headers: authHeaders(),
  });
  communities = communities.filter((entry) => Number(entry.id) !== Number(id));
  render();
  window.refreshPendingCommunitiesCount?.();
}

async function confirmAction(id, action) {
  const community = communities.find((entry) => Number(entry.id) === Number(id));
  const name = community?.name || t("communities.communityFallback");
  const messageKey = action === "approve" ? "admin.confirmApproveCommunity" : "admin.confirmRejectCommunity";
  if (!window.confirm(t(messageKey).replace("{name}", name))) return;
  await setCommunityStatus(id, action);
}

window.addEventListener("DOMContentLoaded", async () => {
  if (!requireAdmin()) return;

  bindLogout();
  highlightActiveNav();

  document.getElementById("refreshCommunitiesBtn")?.addEventListener("click", () =>
    loadCommunities().catch(showPageError),
  );

  document.getElementById("communitiesList")?.addEventListener("click", (event) => {
    const approveBtn = event.target.closest("[data-approve]");
    if (approveBtn) {
      confirmAction(approveBtn.dataset.approve, "approve").catch(showPageError);
      return;
    }
    const rejectBtn = event.target.closest("[data-reject]");
    if (rejectBtn) {
      confirmAction(rejectBtn.dataset.reject, "reject").catch(showPageError);
    }
  });

  window.addEventListener("trendflix:languagechange", render);

  try {
    await loadCommunities();
  } catch (error) {
    showPageError(error);
  }
});
