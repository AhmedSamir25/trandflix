const TOKEN_KEY = "trendflix.token";
const FALLBACK_IMAGE_BASE = "https://placehold.co/500x750/0f172a/f8fafc";

function requireAuth() {
  const token = localStorage.getItem(TOKEN_KEY);
  if (!token) {
    window.location.replace("/pages/auth/auth.html");
    return null;
  }
  return token;
}

function t(key) {
  return window.TrendFlixI18n?.t(key) ?? key;
}

function escapeHtml(s) {
  return String(s ?? "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#039;");
}

function getFallbackImage(title) {
  return `${FALLBACK_IMAGE_BASE}?text=${encodeURIComponent(title || "TrendFlix")}`;
}

function formatDate(dateString) {
  if (!dateString) return "";
  const date = new Date(dateString);
  if (isNaN(date.getTime())) return "";
  return date.toLocaleDateString("en-US", { year: "numeric", month: "long", day: "numeric" });
}

function renderStars(rating, max = 10, count = 5) {
  const normalized = Math.round((rating / max) * count);
  return Array.from({ length: count }, (_, i) =>
    `<span class="star${i < normalized ? " star-on" : ""}">★</span>`
  ).join("");
}

function renderReviewStars(rating, max = 5) {
  return Array.from({ length: max }, (_, i) =>
    `<span class="star${i < rating ? " star-on" : ""}">★</span>`
  ).join("");
}

function getItemIdFromLocation() {
  const params = new URLSearchParams(window.location.search);
  const queryId = params.get("id");
  if (queryId) return queryId;

  const match = window.location.pathname.match(/^\/detail\/([^/]+)$/);
  return match ? decodeURIComponent(match[1]) : "";
}

async function fetchJson(url, token) {
  const headers = { Accept: "application/json" };
  if (token) headers.Authorization = `Bearer ${token}`;
  const res = await fetch(url, { headers });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(data?.msg || `Request failed: ${res.status}`);
  return data;
}

function getTypeMeta(type) {
  const map = {
    movie: { icon: "🎬", labelKey: "detail.typeMovie", actionKey: "common.watch", actionIcon: "▶" },
    game:  { icon: "🎮", labelKey: "detail.typeGame",  actionKey: "common.play",  actionIcon: "🎮" },
    book:  { icon: "📚", labelKey: "detail.typeBook",  actionKey: "common.read",  actionIcon: "📖" },
  };
  return map[type] || { icon: "🎬", labelKey: "detail.typeMovie", actionKey: "common.watch", actionIcon: "▶" };
}

function buildMeta(item) {
  const rows = [];
  if (item.director)   rows.push([t("detail.director"),   item.director]);
  if (item.author)     rows.push([t("detail.author"),     item.author]);
  if (item.developer)  rows.push([t("detail.developer"),  item.developer]);
  if (item.duration)   rows.push([t("detail.duration"),   `${item.duration} ${t("detail.mins")}`]);
  if (item.pages_count) rows.push([t("detail.pages"),     item.pages_count]);
  if (item.platform)   rows.push([t("detail.platform"),   item.platform]);
  if (item.release_date) rows.push([t("detail.releaseDate"), formatDate(item.release_date)]);
  return rows
    .map(([label, value]) => `
      <span class="meta-label">${escapeHtml(label)}</span>
      <span class="meta-value">${escapeHtml(String(value))}</span>
    `)
    .join("");
}

function buildCategories(categories) {
  if (!categories?.length) return "";
  return categories.map(c => `<span class="detail-chip">${escapeHtml(c.name || "")}</span>`).join("");
}

function buildReviews(reviews) {
  if (!reviews?.length) {
    return `<p class="no-reviews">${escapeHtml(t("detail.noReviews"))}</p>`;
  }
  return reviews.map(r => `
    <div class="review-card">
      <div class="review-header">
        <div class="review-stars">${renderReviewStars(r.rating)}</div>
        <div style="display:flex;align-items:center;gap:8px">
          <span class="review-rating-num">${r.rating}/5</span>
          <span class="review-date">${formatDate(r.created_at)}</span>
        </div>
      </div>
      ${r.comment ? `<p class="review-comment">${escapeHtml(r.comment)}</p>` : ""}
    </div>
  `).join("");
}

function buildPage(item, reviews) {
  const meta = getTypeMeta(item.type);
  const safeImg = item.cover_image || getFallbackImage(item.title);
  const safeTitle = escapeHtml(item.title || "");
  const contentLink = String(item.content_link || "").trim();
  const metaHtml = buildMeta(item);
  const cats = buildCategories(item.categories);
  const reviewsHtml = buildReviews(reviews);
  const ratingVal = item.rating ? item.rating.toFixed(1) : "";

  return `
    <div class="detail-backdrop" style="background-image:url('${escapeHtml(safeImg)}')"></div>
    <div class="detail-page">

      <nav class="detail-topbar">
        <button class="detail-back" id="backBtn" type="button">
          ← <span>${escapeHtml(t("detail.back"))}</span>
        </button>
        <span class="detail-brand">TrendFlix</span>
        <div class="lang-menu">
          <button class="lang-trigger" type="button" data-lang-trigger aria-label="${escapeHtml(t("common.language"))}" aria-expanded="false">🌐</button>
          <div class="lang-menu-list" data-lang-menu>
            <button class="lang-option" type="button" data-set-lang="en">${escapeHtml(t("common.english"))}</button>
            <button class="lang-option" type="button" data-set-lang="ar">${escapeHtml(t("common.arabic"))}</button>
          </div>
        </div>
      </nav>

      <div class="detail-hero">
        <div class="detail-poster">
          <img src="${escapeHtml(safeImg)}" alt="${safeTitle}" />
          <div class="poster-type-badge">${meta.icon}</div>
        </div>

        <div class="detail-info">
          <div class="detail-type-badge">${meta.icon} ${escapeHtml(t(meta.labelKey))}</div>
          <h1 class="detail-title">${safeTitle}</h1>

          ${ratingVal ? `
            <div class="detail-rating">
              <div class="rating-stars">${renderStars(item.rating)}</div>
              <span class="rating-num">${escapeHtml(ratingVal)}</span>
              <span class="rating-scale">/ 10</span>
            </div>
          ` : ""}

          ${cats ? `<div class="detail-categories">${cats}</div>` : ""}

          ${metaHtml ? `<div class="detail-meta">${metaHtml}</div>` : ""}

          <div class="detail-actions">
            <button class="action-primary" id="actionBtn" type="button" ${contentLink ? `data-content-link="${escapeHtml(contentLink)}"` : "disabled"}>
              ${meta.actionIcon} ${escapeHtml(t(meta.actionKey))}
            </button>
            <button class="action-fav" id="favBtn" type="button" data-item-id="${item.id}">
              ❤ <span id="favBtnLabel">${escapeHtml(t("detail.addFavorite"))}</span>
            </button>
            <button class="action-save" id="watchLaterBtn" type="button" data-item-id="${item.id}">
              🕒 <span id="watchLaterBtnLabel">${escapeHtml(t("detail.addWatchLater"))}</span>
            </button>
            <button class="action-save" id="saveToListBtn" type="button">
              📋 <span>${escapeHtml(t("detail.saveToList"))}</span>
            </button>
          </div>
        </div>
      </div>

      <div class="hero-divider"></div>

      <div class="detail-sections">
        ${item.description ? `
          <section class="detail-section">
            <h2 class="section-title">${escapeHtml(t("detail.description"))}</h2>
            <p class="detail-description">${escapeHtml(item.description)}</p>
          </section>
        ` : ""}

        <section class="detail-section">
          <h2 class="section-title">${escapeHtml(t("detail.reviews"))}</h2>
          <div class="reviews-grid">${reviewsHtml}</div>
        </section>
      </div>

      <div class="save-to-list-modal" id="saveToListModal" hidden>
        <div class="save-to-list-backdrop" id="saveToListBackdrop"></div>
        <div class="save-to-list-content">
          <div class="save-to-list-header">
            <h3>${escapeHtml(t("detail.saveToList"))}</h3>
            <button class="save-to-list-close" id="saveToListClose" type="button">✕</button>
          </div>
          <form class="create-list-inline" id="createListInline">
            <input type="text" id="newListName" placeholder="${escapeHtml(t("lists.createPlaceholder"))}" required maxlength="100" />
            <button type="submit" id="createListInlineBtn">${escapeHtml(t("lists.createBtn"))}</button>
          </form>
          <div class="save-to-list-items" id="saveToListItems"></div>
        </div>
      </div>

    </div>
  `;
}

function attachHandlers(item, reviews, token) {
  document.getElementById("backBtn")?.addEventListener("click", () => {
    if (document.referrer && new URL(document.referrer).pathname !== window.location.pathname) {
      history.back();
    } else {
      window.location.href = "/pages/app.html";
    }
  });

  document.getElementById("actionBtn")?.addEventListener("click", () => {
    const contentLink = String(item.content_link || "").trim();
    if (!contentLink) return;
    window.open(contentLink, "_blank", "noopener,noreferrer");
  });

  const favBtn = document.getElementById("favBtn");
  if (favBtn) {
    favBtn.addEventListener("click", async () => {
      const isActive = favBtn.classList.contains("active");
      const id = favBtn.getAttribute("data-item-id");
      if (!id) return;

      favBtn.disabled = true;
      try {
        const method = isActive ? "DELETE" : "POST";
        await fetch(`/favorites/${id}`, {
          method,
          headers: { Authorization: `Bearer ${token}`, Accept: "application/json" },
        });
        favBtn.classList.toggle("active", !isActive);
        const label = document.getElementById("favBtnLabel");
        if (label) label.textContent = t(isActive ? "detail.addFavorite" : "detail.removeFavorite");
      } catch {
      } finally {
        favBtn.disabled = false;
      }
    });
  }

  const wlBtn = document.getElementById("watchLaterBtn");
  if (wlBtn) {
    wlBtn.addEventListener("click", async () => {
      const isActive = wlBtn.classList.contains("active");
      const id = wlBtn.getAttribute("data-item-id");
      if (!id) return;

      wlBtn.disabled = true;
      try {
        const method = isActive ? "DELETE" : "POST";
        await fetch(`/watch-later/${id}`, {
          method,
          headers: { Authorization: `Bearer ${token}`, Accept: "application/json" },
        });
        wlBtn.classList.toggle("active", !isActive);
        const label = document.getElementById("watchLaterBtnLabel");
        if (label) label.textContent = t(isActive ? "detail.addWatchLater" : "detail.removeWatchLater");
        if (isActive) watchLaterItemIds.delete(id);
        else watchLaterItemIds.add(id);
      } catch {
      } finally {
        wlBtn.disabled = false;
      }
    });
  }

  const saveToListBtn = document.getElementById("saveToListBtn");
  if (saveToListBtn) {
    saveToListBtn.addEventListener("click", () => {
      openSaveToListModal(token);
    });
  }

  document.getElementById("saveToListClose")?.addEventListener("click", closeSaveToListModal);
  document.getElementById("saveToListBackdrop")?.addEventListener("click", closeSaveToListModal);

  document.getElementById("createListInline")?.addEventListener("submit", async (e) => {
    e.preventDefault();
    const input = document.getElementById("newListName");
    const name = (input?.value || "").trim();
    if (!name) return;

    const btn = document.getElementById("createListInlineBtn");
    btn.disabled = true;
    try {
      const res = await fetch("/lists", {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
          "Content-Type": "application/json",
          Accept: "application/json",
        },
        body: JSON.stringify({ name }),
      });
      const data = await res.json().catch(() => ({}));
      if (res.ok && data?.list) {
        userLists.unshift(data.list);
        listItemIds[data.list.id] = new Set();
        renderSaveToListItems(token, item.id);
        if (input) input.value = "";
      }
    } catch {
    } finally {
      btn.disabled = false;
    }
  });
}

let currentItem = null;
let currentReviews = [];
let currentToken = null;
let watchLaterItemIds = new Set();
let userLists = [];
let listItemIds = {};

function rerender() {
  if (!currentItem) return;
  const root = document.getElementById("detailRoot");
  if (!root) return;
  root.innerHTML = buildPage(currentItem, currentReviews);
  window.TrendFlixI18n?.translatePage();
  attachHandlers(currentItem, currentReviews, currentToken);
  syncWatchLaterState();
  syncSaveToListStates();
}

function syncWatchLaterState() {
  const wlBtn = document.getElementById("watchLaterBtn");
  const label = document.getElementById("watchLaterBtnLabel");
  if (!wlBtn) return;
  const itemId = String(wlBtn.getAttribute("data-item-id") || "");
  const isActive = watchLaterItemIds.has(itemId);
  wlBtn.classList.toggle("active", isActive);
  if (label) label.textContent = t(isActive ? "detail.removeWatchLater" : "detail.addWatchLater");
}

function syncSaveToListStates() {
  userLists.forEach((list) => {
    const itemIds = listItemIds[list.id] || new Set();
    if (itemIds.has(String(currentItem?.id || ""))) {
      const btn = document.getElementById(`listToggle-${list.id}`);
      if (btn) btn.classList.add("active");
    }
  });
}

function openSaveToListModal(token) {
  const modal = document.getElementById("saveToListModal");
  if (modal) {
    modal.hidden = false;
    renderSaveToListItems(token, currentItem?.id);
  }
}

function closeSaveToListModal() {
  const modal = document.getElementById("saveToListModal");
  if (modal) modal.hidden = true;
}

function renderSaveToListItems(token, itemId) {
  const container = document.getElementById("saveToListItems");
  if (!container || !itemId) return;

  if (!userLists.length) {
    container.innerHTML = `<p class="save-to-list-empty">${escapeHtml(t("lists.noListsYet"))}</p>`;
    return;
  }

  container.innerHTML = userLists.map((list) => {
    const itemIds = listItemIds[list.id] || new Set();
    const isActive = itemIds.has(String(itemId));
    const name = escapeHtml(list.name || "");
    const count = itemIds.size;

    return `
      <div class="save-to-list-row" data-list-id="${list.id}">
        <button class="save-to-list-toggle ${isActive ? "active" : ""}" id="listToggle-${list.id}" type="button" data-list-id="${list.id}">
          ${isActive ? "✓" : "○"}
        </button>
        <div class="save-to-list-name">
          <span>${name}</span>
          <span class="save-to-list-count">(${count})</span>
        </div>
        <button class="save-to-list-delete" type="button" data-list-id="${list.id}">🗑</button>
      </div>
    `;
  }).join("");

  container.querySelectorAll(".save-to-list-toggle").forEach((btn) => {
    btn.addEventListener("click", async () => {
      const listId = btn.getAttribute("data-list-id");
      if (!listId) return;
      btn.disabled = true;
      try {
        const itemIds = listItemIds[listId] || new Set();
        const isAdded = itemIds.has(String(itemId));
        const method = isAdded ? "DELETE" : "POST";
        await fetch(`/lists/${listId}/items/${itemId}`, {
          method,
          headers: { Authorization: `Bearer ${token}`, Accept: "application/json" },
        });
        if (isAdded) {
          itemIds.delete(String(itemId));
        } else {
          itemIds.add(String(itemId));
        }
        listItemIds[listId] = itemIds;
        btn.classList.toggle("active", !isAdded);
        btn.innerHTML = isAdded ? "○" : "✓";
        const countEl = btn.closest(".save-to-list-row")?.querySelector(".save-to-list-count");
        if (countEl) countEl.textContent = `(${itemIds.size})`;
      } catch {
      } finally {
        btn.disabled = false;
      }
    });
  });

  container.querySelectorAll(".save-to-list-delete").forEach((btn) => {
    btn.addEventListener("click", async () => {
      const listId = btn.getAttribute("data-list-id");
      if (!listId) return;
      btn.disabled = true;
      try {
        await fetch(`/lists/${listId}`, {
          method: "DELETE",
          headers: { Authorization: `Bearer ${token}`, Accept: "application/json" },
        });
        userLists = userLists.filter((l) => String(l.id) !== listId);
        delete listItemIds[listId];
        renderSaveToListItems(token, itemId);
      } catch {
        btn.disabled = false;
      }
    });
  });
}

window.addEventListener("DOMContentLoaded", async () => {
  const token = requireAuth();
  if (!token) return;
  currentToken = token;

  const id = getItemIdFromLocation();
  if (!id) {
    window.location.replace("/pages/app.html");
    return;
  }

  const root = document.getElementById("detailRoot");

  try {
    const [itemRes, reviewsRes] = await Promise.all([
      fetchJson(`/items/${id}`, token),
      fetchJson(`/reviews/item/${id}`, token).catch(() => ({ reviews: [] })),
    ]);

    const item = itemRes?.item || null;
    if (!item?.id) {
      root.innerHTML = `
        <div class="detail-error">
          <p>${escapeHtml(t("detail.notFound"))}</p>
          <a href="/pages/app.html" class="detail-error-back">← ${escapeHtml(t("detail.back"))}</a>
        </div>`;
      return;
    }

    currentItem = item;
    currentReviews = Array.isArray(reviewsRes?.reviews) ? reviewsRes.reviews : [];

    document.title = `TrendFlix · ${item.title || "Details"}`;

    root.innerHTML = buildPage(item, currentReviews);
    window.TrendFlixI18n?.translatePage();
    attachHandlers(item, currentReviews, token);

    try {
      const [wlRes, listsRes] = await Promise.all([
        fetchJson("/watch-later", token).catch(() => ({ items: [] })),
        fetchJson("/lists", token).catch(() => ({ lists: [] })),
      ]);

      watchLaterItemIds = new Set(
        (Array.isArray(wlRes?.items) ? wlRes.items : []).map((i) => String(i.id)),
      );
      userLists = Array.isArray(listsRes?.lists) ? listsRes.lists : [];
      listItemIds = {};
      userLists.forEach((list) => {
        listItemIds[list.id] = new Set();
      });

      syncWatchLaterState();
    } catch {
    }

    window.addEventListener("trendflix:languagechange", rerender);

  } catch (err) {
    console.error("Failed to load item", err);
    root.innerHTML = `
      <div class="detail-error">
        <p>${escapeHtml(t("detail.loadFailed"))}</p>
        <a href="/pages/app.html" class="detail-error-back">← ${escapeHtml(t("detail.back"))}</a>
      </div>`;
  }
});
