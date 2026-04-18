const TOKEN_KEY = "trendflix.token";
const FALLBACK_IMAGE_BASE = "https://placehold.co/500x700/0f172a/f8fafc";

let favoriteItems = [];
let currentStatusKey = "favorites.loading";

function t(key) {
  return window.TrendFlixI18n?.t(key) ?? key;
}

function requireAuth() {
  const token = localStorage.getItem(TOKEN_KEY);
  if (!token) {
    window.location.replace("/pages/auth/auth.html");
    return null;
  }
  return token;
}

function escapeHtml(value) {
  return String(value ?? "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#039;");
}

function getFallbackImage(title) {
  return `${FALLBACK_IMAGE_BASE}?text=${encodeURIComponent(title || "TrendFlix")}`;
}

function getDetailHref(itemId) {
  return `/pages/detail.html?id=${encodeURIComponent(itemId)}`;
}

function formatType(type) {
  const map = {
    movie: t("detail.typeMovie"),
    game: t("detail.typeGame"),
    book: t("detail.typeBook"),
  };
  return map[type] || type || "";
}

function formatDate(dateString) {
  if (!dateString) return "";
  const date = new Date(dateString);
  if (isNaN(date.getTime())) return "";
  return date.getFullYear();
}

function setStatus(messageKey) {
  currentStatusKey = messageKey;
  const statusEl = document.getElementById("favoritesStatus");
  const gridEl = document.getElementById("favoritesGrid");
  if (statusEl) {
    statusEl.hidden = false;
    statusEl.textContent = t(messageKey);
  }
  if (gridEl) {
    gridEl.hidden = true;
    gridEl.innerHTML = "";
  }
}

function renderFavorites() {
  const statusEl = document.getElementById("favoritesStatus");
  const gridEl = document.getElementById("favoritesGrid");
  if (!statusEl || !gridEl) return;

  if (!favoriteItems.length) {
    setStatus("favorites.empty");
    return;
  }

  statusEl.hidden = true;
  gridEl.hidden = false;
  gridEl.innerHTML = favoriteItems.map((item) => {
    const cover = escapeHtml(item.cover_image || getFallbackImage(item.title));
    const title = escapeHtml(item.title || "");
    const type = escapeHtml(formatType(item.type));
    const year = escapeHtml(String(formatDate(item.release_date) || ""));
    const categories = (item.categories || []).map((category) => escapeHtml(category.name || "")).filter(Boolean).join(" • ");

    return `
      <article class="favorite-card" data-item-id="${item.id}" data-detail-url="${escapeHtml(getDetailHref(item.id))}" tabindex="0" role="link" aria-label="Open details for ${title}">
        <button class="favorite-remove" type="button" data-remove-id="${item.id}">${escapeHtml(t("favorites.remove"))}</button>
        <img class="favorite-cover" src="${cover}" alt="${title}" loading="lazy" />
        <div class="favorite-body">
          <div class="favorite-type">${type}</div>
          <h2 class="favorite-title">${title}</h2>
          <div class="favorite-meta">${year}</div>
          ${categories ? `<div class="favorite-categories">${categories}</div>` : ""}
        </div>
      </article>
    `;
  }).join("");
}

async function fetchJson(url, token, options = {}) {
  const headers = {
    Accept: "application/json",
    ...(options.headers || {}),
  };
  if (token) headers.Authorization = `Bearer ${token}`;

  const response = await fetch(url, { ...options, headers });
  const data = await response.json().catch(() => ({}));
  if (!response.ok) {
    throw new Error(data?.msg || `Request failed: ${response.status}`);
  }

  return data;
}

async function loadFavorites(token) {
  setStatus("favorites.loading");
  const response = await fetchJson("/favorites", token);
  favoriteItems = Array.isArray(response?.items) ? response.items : [];
  renderFavorites();
}

async function removeFavorite(itemId, token) {
  await fetchJson(`/favorites/${itemId}`, token, { method: "DELETE" });
  favoriteItems = favoriteItems.filter((item) => String(item.id) !== String(itemId));
  renderFavorites();
}

function openCardDetail(cardEl) {
  const detailUrl = cardEl?.getAttribute("data-detail-url");
  if (!detailUrl) return;
  window.location.href = detailUrl;
}

function handleLanguageChange() {
  if (favoriteItems.length) {
    renderFavorites();
    return;
  }

  const statusEl = document.getElementById("favoritesStatus");
  if (statusEl) statusEl.textContent = t(currentStatusKey);
}

window.addEventListener("DOMContentLoaded", async () => {
  const token = requireAuth();
  if (!token) return;

  try {
    await loadFavorites(token);
  } catch (error) {
    console.error("Failed to load favorites", error);
    setStatus("favorites.loadFailed");
  }

  document.body.addEventListener("click", async (event) => {
    const removeBtn = event.target.closest?.("[data-remove-id]");
    if (removeBtn) {
      event.stopPropagation();
      removeBtn.disabled = true;
      try {
        await removeFavorite(removeBtn.getAttribute("data-remove-id"), token);
      } catch (error) {
        console.error("Failed to remove favorite", error);
        removeBtn.disabled = false;
      }
      return;
    }

    const cardEl = event.target.closest?.(".favorite-card[data-detail-url]");
    if (cardEl) openCardDetail(cardEl);
  });

  document.body.addEventListener("keydown", (event) => {
    const cardEl = event.target.closest?.(".favorite-card[data-detail-url]");
    if (!cardEl) return;
    if (event.key !== "Enter" && event.key !== " ") return;
    event.preventDefault();
    openCardDetail(cardEl);
  });

  window.addEventListener("trendflix:languagechange", handleLanguageChange);
});
