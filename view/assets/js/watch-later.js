const TOKEN_KEY = "trendflix.token";
const FALLBACK_IMAGE_BASE = "https://placehold.co/500x700/0f172a/f8fafc";

let watchLaterItems = [];
let currentStatusKey = "watchLater.loading";

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
  const statusEl = document.getElementById("listsStatus");
  const gridEl = document.getElementById("listsGrid");
  if (statusEl) {
    statusEl.hidden = false;
    statusEl.textContent = t(messageKey);
  }
  if (gridEl) {
    gridEl.hidden = true;
    gridEl.innerHTML = "";
  }
}

function renderItems() {
  const statusEl = document.getElementById("listsStatus");
  const gridEl = document.getElementById("listsGrid");
  if (!statusEl || !gridEl) return;

  if (!watchLaterItems.length) {
    setStatus("watchLater.empty");
    return;
  }

  statusEl.hidden = true;
  gridEl.hidden = false;
  gridEl.innerHTML = watchLaterItems.map((item) => {
    const cover = escapeHtml(item.cover_image || getFallbackImage(item.title));
    const title = escapeHtml(item.title || "");
    const type = escapeHtml(formatType(item.type));
    const year = escapeHtml(String(formatDate(item.release_date) || ""));
    const categories = (item.categories || []).map((category) => escapeHtml(category.name || "")).filter(Boolean).join(" • ");

    return `
      <article class="lists-card" data-item-id="${item.id}" data-detail-url="${escapeHtml(getDetailHref(item.id))}" tabindex="0" role="link" aria-label="Open details for ${title}">
        <button class="lists-card-remove" type="button" data-remove-id="${item.id}">${escapeHtml(t("watchLater.remove"))}</button>
        <img class="lists-card-cover" src="${cover}" alt="${title}" loading="lazy" />
        <div class="lists-card-body">
          <div class="lists-card-type">${type}</div>
          <h2 class="lists-card-title">${title}</h2>
          <div class="lists-card-meta">${year}</div>
          ${categories ? `<div class="lists-card-categories">${categories}</div>` : ""}
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

async function loadWatchLater(token) {
  setStatus("watchLater.loading");
  const response = await fetchJson("/watch-later", token);
  watchLaterItems = Array.isArray(response?.items) ? response.items : [];
  renderItems();
}

async function removeItem(itemId, token) {
  await fetchJson(`/watch-later/${itemId}`, token, { method: "DELETE" });
  watchLaterItems = watchLaterItems.filter((item) => String(item.id) !== String(itemId));
  renderItems();
}

function openCardDetail(cardEl) {
  const detailUrl = cardEl?.getAttribute("data-detail-url");
  if (!detailUrl) return;
  window.location.href = detailUrl;
}

function handleLanguageChange() {
  if (watchLaterItems.length) {
    renderItems();
    return;
  }

  const statusEl = document.getElementById("listsStatus");
  if (statusEl) statusEl.textContent = t(currentStatusKey);
}

window.addEventListener("DOMContentLoaded", async () => {
  const token = requireAuth();
  if (!token) return;

  try {
    await loadWatchLater(token);
  } catch (error) {
    console.error("Failed to load watch later", error);
    setStatus("watchLater.loadFailed");
  }

  document.body.addEventListener("click", async (event) => {
    const removeBtn = event.target.closest?.("[data-remove-id]");
    if (removeBtn) {
      event.stopPropagation();
      removeBtn.disabled = true;
      try {
        await removeItem(removeBtn.getAttribute("data-remove-id"), token);
      } catch (error) {
        console.error("Failed to remove item", error);
        removeBtn.disabled = false;
      }
      return;
    }

    const cardEl = event.target.closest?.(".lists-card[data-detail-url]");
    if (cardEl) openCardDetail(cardEl);
  });

  document.body.addEventListener("keydown", (event) => {
    const cardEl = event.target.closest?.(".lists-card[data-detail-url]");
    if (!cardEl) return;
    if (event.key !== "Enter" && event.key !== " ") return;
    event.preventDefault();
    openCardDetail(cardEl);
  });

  window.addEventListener("trendflix:languagechange", handleLanguageChange);
});
