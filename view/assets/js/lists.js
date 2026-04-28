const TOKEN_KEY = "trendflix.token";
const FALLBACK_IMAGE_BASE = "https://placehold.co/500x700/0f172a/f8fafc";

let userLists = [];
let currentList = null;
let currentListItems = [];
let currentStatusKey = "lists.loading";

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
  const actionsEl = document.getElementById("listsActions");
  if (statusEl) {
    statusEl.hidden = false;
    statusEl.textContent = t(messageKey);
  }
  if (gridEl) {
    gridEl.hidden = true;
    gridEl.innerHTML = "";
  }
  if (actionsEl) actionsEl.hidden = true;
}

function renderLists() {
  const statusEl = document.getElementById("listsStatus");
  const gridEl = document.getElementById("listsGrid");
  const actionsEl = document.getElementById("listsActions");
  const listDetail = document.getElementById("listDetail");
  if (!statusEl || !gridEl) return;

  if (listDetail) listDetail.hidden = true;

  if (!userLists.length) {
    setStatus("lists.empty");
    if (actionsEl) actionsEl.hidden = false;
    return;
  }

  statusEl.hidden = true;
  if (actionsEl) actionsEl.hidden = false;
  gridEl.hidden = false;
  gridEl.innerHTML = userLists.map((list) => {
    const name = escapeHtml(list.name || "");
    return `
      <article class="lists-card list-card" data-list-id="${list.id}" tabindex="0" role="link" aria-label="Open list ${name}">
        <div class="lists-card-body">
          <div class="list-card-icon">📁</div>
          <h2 class="lists-card-title">${name}</h2>
          <div class="list-card-count">${t("lists.itemCount")}: 0</div>
        </div>
      </article>
    `;
  }).join("");
}

function renderListDetail(list, items) {
  const statusEl = document.getElementById("listsStatus");
  const gridEl = document.getElementById("listsGrid");
  const actionsEl = document.getElementById("listsActions");
  const listDetail = document.getElementById("listDetail");
  const listDetailTitle = document.getElementById("listDetailTitle");
  const listDetailStatus = document.getElementById("listDetailStatus");
  const listDetailGrid = document.getElementById("listDetailGrid");
  if (!statusEl || !gridEl || !listDetail) return;

  statusEl.hidden = true;
  gridEl.hidden = true;
  if (actionsEl) actionsEl.hidden = true;
  listDetail.hidden = false;

  if (listDetailTitle) listDetailTitle.textContent = list.name;

  if (!items.length) {
    if (listDetailStatus) {
      listDetailStatus.hidden = false;
      listDetailStatus.textContent = t("lists.emptyList");
    }
    if (listDetailGrid) {
      listDetailGrid.hidden = true;
      listDetailGrid.innerHTML = "";
    }
    return;
  }

  if (listDetailStatus) listDetailStatus.hidden = true;
  if (listDetailGrid) {
    listDetailGrid.hidden = false;
    listDetailGrid.innerHTML = items.map((item) => {
      const cover = escapeHtml(item.cover_image || getFallbackImage(item.title));
      const title = escapeHtml(item.title || "");
      const type = escapeHtml(formatType(item.type));
      const year = escapeHtml(String(formatDate(item.release_date) || ""));
      const categories = (item.categories || []).map((category) => escapeHtml(category.name || "")).filter(Boolean).join(" • ");

      return `
        <article class="lists-card" data-item-id="${item.id}" data-detail-url="${escapeHtml(getDetailHref(item.id))}" tabindex="0" role="link" aria-label="Open details for ${title}">
          <button class="lists-card-remove" type="button" data-remove-id="${item.id}">${escapeHtml(t("lists.removeFromList"))}</button>
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

async function loadLists(token) {
  setStatus("lists.loading");
  const response = await fetchJson("/lists", token);
  userLists = Array.isArray(response?.lists) ? response.lists : [];
  renderLists();
}

async function createList(token, name) {
  const response = await fetchJson("/lists", token, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ name }),
  });
  if (response?.list) {
    userLists.unshift(response.list);
    renderLists();
  }
}

async function loadListDetail(token, listId) {
  const response = await fetchJson(`/lists/${listId}`, token);
  currentList = response?.list || null;
  currentListItems = Array.isArray(response?.items) ? response.items : [];
  renderListDetail(currentList, currentListItems);
}

async function deleteList(token, listId) {
  await fetchJson(`/lists/${listId}`, token, { method: "DELETE" });
  userLists = userLists.filter((l) => String(l.id) !== String(listId));
  renderLists();
}

async function removeItemFromList(token, listId, itemId) {
  await fetchJson(`/lists/${listId}/items/${itemId}`, token, { method: "DELETE" });
  currentListItems = currentListItems.filter((item) => String(item.id) !== String(itemId));
  renderListDetail(currentList, currentListItems);
}

function openCardDetail(cardEl) {
  const detailUrl = cardEl?.getAttribute("data-detail-url");
  if (!detailUrl) return;
  window.location.href = detailUrl;
}

function handleLanguageChange() {
  if (currentList) {
    renderListDetail(currentList, currentListItems);
    return;
  }

  if (userLists.length) {
    renderLists();
    return;
  }

  const statusEl = document.getElementById("listsStatus");
  if (statusEl) statusEl.textContent = t(currentStatusKey);
}

window.addEventListener("DOMContentLoaded", async () => {
  const token = requireAuth();
  if (!token) return;

  try {
    await loadLists(token);
  } catch (error) {
    console.error("Failed to load lists", error);
    setStatus("lists.loadFailed");
  }

  document.getElementById("backToLists")?.addEventListener("click", () => {
    currentList = null;
    currentListItems = [];
    renderLists();
  });

  document.getElementById("deleteListBtn")?.addEventListener("click", async () => {
    if (!currentList) return;
    const btn = document.getElementById("deleteListBtn");
    btn.disabled = true;
    try {
      await deleteList(token, currentList.id);
    } catch (error) {
      console.error("Failed to delete list", error);
      btn.disabled = false;
    }
  });

  document.getElementById("createListForm")?.addEventListener("submit", async (e) => {
    e.preventDefault();
    const input = document.getElementById("listNameInput");
    const name = (input?.value || "").trim();
    if (!name) return;

    const btn = document.getElementById("createListBtn");
    btn.disabled = true;
    try {
      await createList(token, name);
      if (input) input.value = "";
    } catch (error) {
      console.error("Failed to create list", error);
    } finally {
      btn.disabled = false;
    }
  });

  document.body.addEventListener("click", async (event) => {
    const removeBtn = event.target.closest?.("[data-remove-id]");
    if (removeBtn && currentList) {
      event.stopPropagation();
      removeBtn.disabled = true;
      try {
        await removeItemFromList(token, currentList.id, removeBtn.getAttribute("data-remove-id"));
      } catch (error) {
        console.error("Failed to remove item", error);
        removeBtn.disabled = false;
      }
      return;
    }

    const listCard = event.target.closest?.(".list-card[data-list-id]");
    if (listCard) {
      event.stopPropagation();
      const listId = listCard.getAttribute("data-list-id");
      if (listId) {
        try {
          await loadListDetail(token, listId);
        } catch (error) {
          console.error("Failed to load list detail", error);
        }
      }
      return;
    }

    const cardEl = event.target.closest?.(".lists-card[data-detail-url]");
    if (cardEl) openCardDetail(cardEl);
  });

  document.body.addEventListener("keydown", (event) => {
    const listCard = event.target.closest?.(".list-card[data-list-id]");
    if (listCard) {
      if (event.key !== "Enter" && event.key !== " ") return;
      event.preventDefault();
      const listId = listCard.getAttribute("data-list-id");
      if (listId) {
        loadListDetail(token, listId).catch(console.error);
      }
      return;
    }

    const cardEl = event.target.closest?.(".lists-card[data-detail-url]");
    if (!cardEl) return;
    if (event.key !== "Enter" && event.key !== " ") return;
    event.preventDefault();
    openCardDetail(cardEl);
  });

  window.addEventListener("trendflix:languagechange", handleLanguageChange);
});
