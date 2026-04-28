const TOKEN_KEY = "trendflix.token";

const SECTION_ORDER = ["movie", "game", "book"];

const SECTION_META = {
  movie: { icon: "🎬", titleKey: "app.movies" },
  game: { icon: "🎮", titleKey: "app.games" },
  book: { icon: "📚", titleKey: "app.books" },
};

const FALLBACK_IMAGE_BASE = "https://placehold.co/500x700/0f172a/f8fafc";
const CHAT_HISTORY_LIMIT = 8;
const BANNER_LOG_PREFIX = "[TrendFlix banner]";
const BANNER_REQUEST_TIMEOUT_MS = 8000;
const BANNER_ROTATION_INTERVAL_MS = 6000;

let items = [];
let categories = [];
let homeBanners = [];
let activeBannerIndex = 0;
let bannerRotationTimer = 0;
let searchQuery = "";
let catalogStatusKey = "app.loadingCatalog";
let currentToken = "";
let favoriteItemIds = new Set();
let chatHistory = [];
let chatPending = false;

const activeCategoryByType = {
  movie: "all",
  game: "all",
  book: "all",
};

function t(key) {
  return window.TrendFlixI18n?.t(key) ?? key;
}

function getLang() {
  return window.TrendFlixI18n?.getLang?.() || "en";
}

function requireAuth() {
  const token = localStorage.getItem(TOKEN_KEY);
  if (!token) {
    window.location.replace("/pages/auth/auth.html");
    return null;
  }
  return token;
}

function parseJwtPayload(token) {
  try {
    const payload = token.split(".")[1] || "";
    const normalized = payload.replaceAll("-", "+").replaceAll("_", "/");
    const padded = normalized.padEnd(Math.ceil(normalized.length / 4) * 4, "=");
    return JSON.parse(window.atob(padded));
  } catch {
    return null;
  }
}

function getCurrentRole() {
  const token = localStorage.getItem(TOKEN_KEY);
  if (!token) return "";

  const payload = parseJwtPayload(token);
  return String(payload?.role || "").trim().toLowerCase();
}

function syncAdminNavLink() {
  const adminNavLink = document.getElementById("adminNavLink");
  if (!adminNavLink) return;

  adminNavLink.hidden = getCurrentRole() !== "admin";
}

function escapeHtml(s) {
  return String(s)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#039;");
}

function getFallbackImage(title) {
  return `${FALLBACK_IMAGE_BASE}?text=${encodeURIComponent(title || "TrendFlix")}`;
}

function getItemsByType(type) {
  return items.filter((item) => item.type === type);
}

function getCategoriesForType(type) {
  const categoryIds = new Set();

  for (const item of getItemsByType(type)) {
    for (const category of item.categories || []) {
      if (category?.id) {
        categoryIds.add(category.id);
      }
    }
  }

  return categories.filter((category) => categoryIds.has(category.id));
}

function getFilteredItems(type) {
  const activeCategory = activeCategoryByType[type] || "all";
  const query = searchQuery.trim().toLowerCase();

  return getItemsByType(type).filter((item) => {
    const matchesCategory =
      activeCategory === "all" || (item.categories || []).some((category) => String(category.id) === activeCategory);
    const matchesSearch = !query || String(item.title || "").toLowerCase().includes(query);
    return matchesCategory && matchesSearch;
  });
}

function formatDate(dateString) {
  if (!dateString) return "";
  const date = new Date(dateString);
  if (isNaN(date.getTime())) return "";
  const year = date.getFullYear();
  const month = date.getMonth() + 1;
  return `${year}-${month}`;
}

function getDetailHref(itemId) {
  return `/pages/detail.html?id=${encodeURIComponent(itemId)}`;
}

function logBanner(message, details) {
  if (typeof details === "undefined") {
    console.log(BANNER_LOG_PREFIX, message);
    return;
  }
  console.log(BANNER_LOG_PREFIX, message, details);
}

function warnBanner(message, details) {
  if (typeof details === "undefined") {
    console.warn(BANNER_LOG_PREFIX, message);
    return;
  }
  console.warn(BANNER_LOG_PREFIX, message, details);
}

function errorBanner(message, details) {
  if (typeof details === "undefined") {
    console.error(BANNER_LOG_PREFIX, message);
    return;
  }
  console.error(BANNER_LOG_PREFIX, message, details);
}

function normalizeBanner(banner) {
  const title = String(banner?.title || "").trim();
  const subtitle = String(banner?.subtitle || "").trim();
  const imageUrl = String(banner?.image_url || "").trim();
  if (!title || !imageUrl) {
    warnBanner("normalize skipped invalid banner", banner);
    return null;
  }

  const normalized = {
    id: banner?.id ?? title,
    title,
    subtitle,
    imageUrl,
  };

  logBanner("normalize success", normalized);
  return normalized;
}

function handleBannerImageLoad(event) {
  const img = event?.currentTarget;
  logBanner("image loaded", {
    src: img?.currentSrc || img?.src || "",
    naturalWidth: img?.naturalWidth || 0,
    naturalHeight: img?.naturalHeight || 0,
  });
}

function handleBannerImageError(event) {
  const img = event?.currentTarget;
  errorBanner("image failed", {
    src: img?.currentSrc || img?.src || "",
  });
  img?.classList.add("is-broken");
}

function createBannerMarkup(banner, index) {
  return `
    <article class="banner-slide${index === activeBannerIndex ? " active" : ""}" aria-hidden="${index === activeBannerIndex ? "false" : "true"}">
      <img class="banner-image" src="${escapeHtml(banner.imageUrl)}" alt="${escapeHtml(banner.title)}" loading="eager" referrerpolicy="no-referrer" onload="handleBannerImageLoad(event)" onerror="handleBannerImageError(event)" />
      <div class="banner-content">
        <h1>${escapeHtml(banner.title)}</h1>
        ${banner.subtitle ? `<p class="banner-description">${escapeHtml(banner.subtitle)}</p>` : ""}
      </div>
    </article>
  `;
}

function createEmptyBannerMarkup() {
  return `
    <article class="banner-slide active" aria-hidden="false">
      <div class="banner-content">
        <h1>${escapeHtml(t("app.bannerFallbackTitle"))}</h1>
        <p class="banner-description">${escapeHtml(t("app.bannerFallbackDescription"))}</p>
      </div>
    </article>
  `;
}

function setBannerLoading(isLoading) {
  const bannerHero = document.getElementById("bannerHero");
  if (!bannerHero) return null;

  bannerHero.classList.toggle("is-loading", isLoading);
  bannerHero.setAttribute("aria-busy", isLoading ? "true" : "false");
  return bannerHero;
}

function clearBannerRotation() {
  if (!bannerRotationTimer) return;
  window.clearInterval(bannerRotationTimer);
  bannerRotationTimer = 0;
}

function showBannerSlide(index) {
  const slides = Array.from(document.querySelectorAll("#bannerHero .banner-slide"));
  if (!slides.length) return;

  activeBannerIndex = ((index % slides.length) + slides.length) % slides.length;

  slides.forEach((slide, slideIndex) => {
    const isActive = slideIndex === activeBannerIndex;
    slide.classList.toggle("active", isActive);
    slide.setAttribute("aria-hidden", isActive ? "false" : "true");
  });
}

function startBannerRotation() {
  clearBannerRotation();
  if (homeBanners.length < 2) return;

  bannerRotationTimer = window.setInterval(() => {
    showBannerSlide(activeBannerIndex + 1);
  }, BANNER_ROTATION_INTERVAL_MS);
}

function renderBanners() {
  const bannerHero = setBannerLoading(false);
  if (!bannerHero) {
    errorBanner("render aborted: #bannerHero not found");
    return;
  }

  clearBannerRotation();

  if (!homeBanners.length) {
    warnBanner("render with empty banner");
    bannerHero.innerHTML = `
      <div class="banner-track">
        ${createEmptyBannerMarkup()}
      </div>
    `;
    return;
  }

  activeBannerIndex = Math.min(activeBannerIndex, homeBanners.length - 1);

  logBanner("render start", {
    count: homeBanners.length,
    activeBannerIndex,
  });

  bannerHero.innerHTML = `
    <div class="banner-track">
      ${homeBanners.map((banner, index) => createBannerMarkup(banner, index)).join("")}
    </div>
  `;

  showBannerSlide(activeBannerIndex);
  startBannerRotation();

  logBanner("render done", {
    htmlLength: bannerHero.innerHTML.length,
    count: homeBanners.length,
  });
}

function card(item) {
  const favoriteLabel = t("app.toggleFavorite");
  const safeName = escapeHtml(item.title || "");
  const safeImg = escapeHtml(item.cover_image || getFallbackImage(item.title));
  const rating = item.rating ? `⭐ ${item.rating}` : "";
  const releaseDate = formatDate(item.release_date);
  const isFavorite = favoriteItemIds.has(String(item.id));
  const categoryNames = (item.categories || [])
    .slice(0, 2)
    .map((c) => escapeHtml(c.name || ""))
    .filter(Boolean)
    .join(" • ");

  return `
    <article
      class="card-item"
      data-detail-url="${escapeHtml(getDetailHref(item.id))}"
      tabindex="0"
      role="link"
      aria-label="Open details for ${safeName}"
    >
      <button class="heart-btn${isFavorite ? " active" : ""}" type="button" aria-label="${escapeHtml(favoriteLabel)}" data-fav data-item-id="${item.id}">❤</button>
      <img src="${safeImg}" alt="${safeName}" loading="lazy" />
      <div class="title">${safeName}</div>
      <div class="card-info">
        <div class="info-row">
          ${rating ? `<span class="info-tag rating">${rating}</span>` : ""}
          ${releaseDate ? `<span class="info-tag date">${releaseDate}</span>` : ""}
        </div>
        ${categoryNames ? `<span class="info-tag categories">${categoryNames}</span>` : ""}
      </div>
    </article>
  `;
}

function createCategoryChips(type) {
  const typeCategories = getCategoriesForType(type);
  const activeCategory = activeCategoryByType[type] || "all";
  const chips = [
    `<button class="${activeCategory === "all" ? "chip active" : "chip"}" data-category-id="all" data-type="${type}" type="button">${escapeHtml(t("app.all"))}</button>`,
  ];

  for (const category of typeCategories) {
    const isActive = String(category.id) === activeCategory;
    chips.push(
      `<button class="${isActive ? "chip active" : "chip"}" data-category-id="${category.id}" data-type="${type}" type="button">${escapeHtml(category.name)}</button>`,
    );
  }

  return chips.join("");
}

function createSection(type) {
  const meta = SECTION_META[type];
  if (!meta) return "";

  const filteredItems = getFilteredItems(type);
  const content = filteredItems.length
    ? filteredItems.map((item) => card(item)).join("")
    : `<p class="row-status">${escapeHtml(t("app.noItemsFound"))}</p>`;

  return `
    <section class="section" data-section="${type}">
      <h2>${meta.icon} <span>${escapeHtml(t(meta.titleKey))}</span></h2>
      <div class="cat-row" data-cat-row="${type}">
        ${createCategoryChips(type)}
      </div>
      <div class="row">${content}</div>
    </section>
  `;
}

function setCatalogStatus(messageKey) {
  catalogStatusKey = messageKey;
  const catalogSections = document.getElementById("catalogSections");
  if (!catalogSections) return;

  catalogSections.innerHTML = `<p class="catalog-status">${escapeHtml(t(messageKey))}</p>`;
}

function renderCatalog() {
  const catalogSections = document.getElementById("catalogSections");
  if (!catalogSections) return;

  const availableTypes = SECTION_ORDER.filter((type) => getItemsByType(type).length > 0);
  if (!availableTypes.length) {
    setCatalogStatus("app.emptyCatalog");
    return;
  }

  catalogStatusKey = "";
  catalogSections.innerHTML = availableTypes.map((type) => createSection(type)).join("");
}

async function fetchJson(url, options = {}, token = "", timeoutMs = 20000) {
  const headers = {
    Accept: "application/json",
    ...(options.headers || {}),
  };
  if (token) headers.Authorization = `Bearer ${token}`;

  const controller = new AbortController();
  const timeoutId = window.setTimeout(() => controller.abort(), timeoutMs);

  try {
    const response = await fetch(url, {
      ...options,
      headers,
      signal: controller.signal,
    });

    const data = await response.json().catch(() => ({}));
    if (!response.ok) {
      throw new Error(data?.msg || `Request failed: ${response.status}`);
    }

    return data;
  } catch (error) {
    if (error?.name === "AbortError") {
      throw new Error("Request timed out");
    }
    throw error;
  } finally {
    window.clearTimeout(timeoutId);
  }
}

async function loadCatalog() {
  setCatalogStatus("app.loadingCatalog");

  const [itemsResponse, categoriesResponse] = await Promise.all([fetchJson("/items"), fetchJson("/categories")]);

  items = Array.isArray(itemsResponse?.items) ? itemsResponse.items : [];
  categories = Array.isArray(categoriesResponse?.categories) ? categoriesResponse.categories : [];

  renderCatalog();
}

async function loadBanners() {
  logBanner("load start");
  setBannerLoading(true);

  try {
    const response = await fetchJson("/banners", {}, "", BANNER_REQUEST_TIMEOUT_MS);
    logBanner("raw response", response);

    homeBanners = (Array.isArray(response?.banners) ? response.banners : [])
      .map(normalizeBanner)
      .filter(Boolean);
    activeBannerIndex = 0;

    if (!homeBanners.length) {
      warnBanner("no valid active banner found in response");
    } else {
      logBanner("selected banners", {
        count: homeBanners.length,
        titles: homeBanners.map((banner) => banner.title),
      });
    }
  } catch (error) {
    errorBanner("load failed", {
      message: error?.message || String(error),
    });
    homeBanners = [];
  } finally {
    renderBanners();
  }
}

async function loadFavoriteItemIds(token) {
  const response = await fetchJson("/favorites", {}, token);
  favoriteItemIds = new Set(
    (Array.isArray(response?.items) ? response.items : []).map((item) => String(item.id)),
  );
}

async function toggleFavorite(btn) {
  const itemId = btn.getAttribute("data-item-id") || "";
  if (!itemId || !currentToken) return;

  const isActive = btn.classList.contains("active");
  btn.disabled = true;

  try {
    await fetchJson(`/favorites/${itemId}`, { method: isActive ? "DELETE" : "POST" }, currentToken);
    btn.classList.toggle("active", !isActive);
    if (isActive) favoriteItemIds.delete(itemId);
    else favoriteItemIds.add(itemId);
  } catch (error) {
    console.error("Failed to toggle favorite", error);
  } finally {
    btn.disabled = false;
  }
}

function openSidebar() {
  document.getElementById("sidebar")?.classList.add("active");
  const overlay = document.getElementById("overlay");
  if (overlay) overlay.style.display = "block";
}

function closeSidebar() {
  document.getElementById("sidebar")?.classList.remove("active");
  const overlay = document.getElementById("overlay");
  if (overlay) overlay.style.display = "none";
}

function toggleSidebar() {
  const sidebar = document.getElementById("sidebar");
  if (!sidebar) return;
  if (sidebar.classList.contains("active")) closeSidebar();
  else openSidebar();
}

function toggleChat(force) {
  const chat = document.getElementById("chatBox");
  if (!chat) return;
  const open = typeof force === "boolean" ? force : chat.style.display !== "flex";
  chat.style.display = open ? "flex" : "none";
  chat.setAttribute("aria-hidden", open ? "false" : "true");
  if (open) {
    document.getElementById("userInput")?.focus();
  }
}

function addMsg(kind, html) {
  const logs = document.getElementById("chatLogs");
  if (!logs) return;
  const div = document.createElement("div");
  div.className = `msg ${kind === "user" ? "user-msg" : "bot-msg"}`;
  div.innerHTML = html;
  logs.appendChild(div);
  logs.scrollTop = logs.scrollHeight;
}

function formatChatMessage(text) {
  return escapeHtml(text).replaceAll("\n", "<br />");
}

function setChatPendingState(isPending) {
  chatPending = isPending;
  const input = document.getElementById("userInput");
  const submit = document.getElementById("chatSubmit");
  if (input) input.disabled = isPending;
  if (submit) submit.disabled = isPending;
}

function pushChatHistory(role, content) {
  chatHistory.push({ role, content });
  if (chatHistory.length > CHAT_HISTORY_LIMIT) {
    chatHistory = chatHistory.slice(-CHAT_HISTORY_LIMIT);
  }
}

async function requestTrendFlixReply(message) {
  const response = await fetchJson(
    "/chat/trendflix",
    {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        message,
        history: chatHistory,
      }),
    },
    currentToken,
  );

  return String(response?.reply || "").trim();
}

function updateSearch(value) {
  searchQuery = value;
  renderCatalog();
}

function handleLanguageChange() {
  renderBanners();

  if (catalogStatusKey) {
    setCatalogStatus(catalogStatusKey);
    return;
  }

  renderCatalog();
}

function openCardDetail(cardEl) {
  const detailUrl = cardEl?.getAttribute("data-detail-url");
  if (!detailUrl) return;
  window.location.href = detailUrl;
}

function logout() {
  localStorage.removeItem(TOKEN_KEY);
  window.location.replace("/pages/auth/auth.html");
}

window.addEventListener("DOMContentLoaded", async () => {
  logBanner("DOMContentLoaded fired");

  const token = requireAuth();
  if (!token) {
    warnBanner("auth token missing, redirecting to login");
    return;
  }
  currentToken = token;

  logBanner("auth token found", {
    tokenLength: token.length,
    hasBannerMount: Boolean(document.getElementById("bannerHero")),
  });

  syncAdminNavLink();

  loadBanners();

  try {
    await loadCatalog();
  } catch (error) {
    console.error("Failed to load catalog", error);
    setCatalogStatus("app.catalogLoadFailed");
  }

  try {
    await loadFavoriteItemIds(token);
    renderCatalog();
  } catch (error) {
    console.error("Failed to load favorites", error);
  }

  document.getElementById("menuBtn")?.addEventListener("click", toggleSidebar);
  document.getElementById("overlay")?.addEventListener("click", closeSidebar);
  document.getElementById("logoutBtn")?.addEventListener("click", logout);

  document.getElementById("searchInput")?.addEventListener("input", (e) => updateSearch(e.target.value || ""));

  document.body.addEventListener("click", async (e) => {
    const chip = e.target.closest?.(".chip");
    if (chip) {
      const type = chip.getAttribute("data-type") || "";
      const categoryId = chip.getAttribute("data-category-id") || "all";
      if (type) {
        activeCategoryByType[type] = categoryId;
        renderCatalog();
      }
      return;
    }

    const fav = e.target.closest?.("[data-fav]");
    if (fav) {
      e.preventDefault();
      e.stopPropagation();
      await toggleFavorite(fav);
      return;
    }

    const cardEl = e.target.closest?.(".card-item[data-detail-url]");
    if (cardEl) {
      openCardDetail(cardEl);
    }
  });

  document.body.addEventListener("keydown", (e) => {
    const cardEl = e.target.closest?.(".card-item[data-detail-url]");
    if (!cardEl) return;
    if (e.key !== "Enter" && e.key !== " ") return;
    e.preventDefault();
    openCardDetail(cardEl);
  });

  document.getElementById("aiToggle")?.addEventListener("click", () => toggleChat());
  document.getElementById("chatClose")?.addEventListener("click", () => toggleChat(false));
  document.getElementById("chatForm")?.addEventListener("submit", async (e) => {
    e.preventDefault();
    if (chatPending) return;

    const input = document.getElementById("userInput");
    const text = (input?.value || "").trim();
    if (!text) return;

    addMsg("user", escapeHtml(text));
    if (input) input.value = "";

    const logs = document.getElementById("chatLogs");
    const loadingMsg = document.createElement("div");
    loadingMsg.className = "msg bot-msg is-loading";
    loadingMsg.innerHTML = escapeHtml(t("app.chatThinking"));
    logs?.appendChild(loadingMsg);
    if (logs) logs.scrollTop = logs.scrollHeight;

    setChatPendingState(true);

    try {
      const reply = await requestTrendFlixReply(text);
      loadingMsg.remove();

      const safeReply = reply || t("app.chatError");
      addMsg("bot", formatChatMessage(safeReply));
      pushChatHistory("user", text);
      pushChatHistory("assistant", safeReply);
    } catch (error) {
      console.error("Failed to fetch TrendFlix chat reply", error);
      loadingMsg.remove();
      addMsg("bot", escapeHtml(t("app.chatError")));
    } finally {
      setChatPendingState(false);
      input?.focus();
    }
  });

  window.addEventListener("trendflix:languagechange", handleLanguageChange);
});
