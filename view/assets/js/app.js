const TOKEN_KEY = "trendflix.token";

const SECTION_ORDER = ["movie", "tv_show", "game", "book"];

const SECTION_META = {
  movie: { icon: "🎬", titleKey: "app.movies" },
  tv_show: { icon: "📺", titleKey: "app.tvShows" },
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
let recommendations = [];
let activeRecommendationType = "all";
let chatHistory = [];
let chatPending = false;

const activeCategoryByType = {
  movie: "all",
  tv_show: "all",
  game: "all",
  book: "all",
};

function t(key) {
  return window.TrendFlixI18n?.t(key) ?? key;
}

function getLang() {
  return window.TrendFlixI18n?.getLang?.() || "en";
}

function categoryName(category) {
  return window.TrendFlixI18n?.categoryName?.(category) || String(category?.name || "");
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

function localizedBannerText(primary, alternate) {
  return window.TrendFlixI18n?.localizedText?.(primary, alternate) ?? String(primary || alternate || "").trim();
}

function normalizeBanner(banner) {
  const title = String(banner?.title || "").trim();
  const titleAr = String(banner?.title_ar || "").trim();
  const subtitle = String(banner?.subtitle || "").trim();
  const subtitleAr = String(banner?.subtitle_ar || "").trim();
  const imageUrl = String(banner?.image_url || "").trim();
  if (!title || !imageUrl) {
    warnBanner("normalize skipped invalid banner", banner);
    return null;
  }

  const normalized = {
    id: banner?.id ?? title,
    title,
    title_ar: titleAr,
    subtitle,
    subtitle_ar: subtitleAr,
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
  const title = localizedBannerText(banner.title, banner.title_ar);
  const subtitle = localizedBannerText(banner.subtitle, banner.subtitle_ar);

  return `
    <article class="banner-slide${index === activeBannerIndex ? " active" : ""}" aria-hidden="${index === activeBannerIndex ? "false" : "true"}">
      <img class="banner-image" src="${escapeHtml(banner.imageUrl)}" alt="${escapeHtml(title)}" loading="eager" referrerpolicy="no-referrer" onload="handleBannerImageLoad(event)" onerror="handleBannerImageError(event)" />
      <div class="banner-content">
        <h1>${escapeHtml(title)}</h1>
        ${subtitle ? `<p class="banner-description">${escapeHtml(subtitle)}</p>` : ""}
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
    .map((c) => escapeHtml(categoryName(c)))
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
      `<button class="${isActive ? "chip active" : "chip"}" data-category-id="${category.id}" data-type="${type}" type="button">${escapeHtml(categoryName(category))}</button>`,
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
      <div class="row-wrapper">
        <button class="row-arrow row-arrow-left" type="button" aria-label="Scroll left">◀</button>
        <div class="row">${content}</div>
        <button class="row-arrow row-arrow-right" type="button" aria-label="Scroll right">▶</button>
      </div>
    </section>
  `;
}

function recommendationCard(rec) {
  const favoriteLabel = t("app.toggleFavorite");
  const safeName = escapeHtml(rec.title || "");
  const safeImg = escapeHtml(rec.cover_image || getFallbackImage(rec.title));
  const rating = rec.rating ? `⭐ ${rec.rating}` : "";
  const isFavorite = favoriteItemIds.has(String(rec.id));
  const categoryNames = (rec.categories || [])
    .slice(0, 2)
    .map((category) => escapeHtml(typeof category === "string" ? category : categoryName(category)))
    .filter(Boolean)
    .join(" • ");

  return `
    <article
      class="card-item card-rec"
      data-detail-url="${escapeHtml(getDetailHref(rec.id))}"
      tabindex="0"
      role="link"
      aria-label="Open details for ${safeName}"
    >
      <button class="heart-btn${isFavorite ? " active" : ""}" type="button" aria-label="${escapeHtml(favoriteLabel)}" data-fav data-item-id="${rec.id}">❤</button>
      <img src="${safeImg}" alt="${safeName}" loading="lazy" />
      <div class="title">${safeName}</div>
      <div class="card-info">
        ${rating ? `<span class="info-tag rating">${rating}</span>` : ""}
        ${categoryNames ? `<span class="info-tag categories">${categoryNames}</span>` : ""}
      </div>
    </article>
  `;
}

function getFilteredRecommendations() {
  if (activeRecommendationType === "all") {
    return recommendations;
  }
  return recommendations.filter((rec) => rec.type === activeRecommendationType);
}

function createRecommendationTypeChips() {
  const chips = [
    `<button class="${activeRecommendationType === "all" ? "chip active" : "chip"}" data-rec-type="all" type="button">${escapeHtml(t("app.all"))}</button>`,
  ];

  for (const type of SECTION_ORDER) {
    const meta = SECTION_META[type];
    if (!meta) continue;
    chips.push(
      `<button class="${activeRecommendationType === type ? "chip active" : "chip"}" data-rec-type="${type}" type="button">${meta.icon} ${escapeHtml(t(meta.titleKey))}</button>`,
    );
  }

  return chips.join("");
}

function createRecommendationSection() {
  if (!recommendations.length) return "";
  if (searchQuery.trim()) return "";

  const filteredRecommendations = getFilteredRecommendations();
  const content = filteredRecommendations.length
    ? filteredRecommendations.map((rec) => recommendationCard(rec)).join("")
    : `<p class="row-status">${escapeHtml(t("app.noItemsFound"))}</p>`;

  return `
    <section class="section section-rec" data-section="recommendations">
      <h2>✨ <span>${escapeHtml(t("app.recommendedForYou"))}</span></h2>
      <div class="cat-row rec-type-row" data-rec-type-row>
        ${createRecommendationTypeChips()}
      </div>
      <div class="row-wrapper">
        <button class="row-arrow row-arrow-left" type="button" aria-label="Scroll left">◀</button>
        <div class="row">${content}</div>
        <button class="row-arrow row-arrow-right" type="button" aria-label="Scroll right">▶</button>
      </div>
    </section>
  `;
}

function normalizeRecommendationFromItem(item, reason) {
  return {
    id: item.id,
    title: item.title,
    type: item.type,
    cover_image: item.cover_image,
    rating: item.rating,
    categories: (item.categories || []).map((category) => categoryName(category)).filter(Boolean),
    reason,
  };
}

function buildLocalRecommendationFallback(limit = 20) {
  return items
    .filter((item) => item?.id && !favoriteItemIds.has(String(item.id)))
    .slice()
    .sort((a, b) => Number(b.rating || 0) - Number(a.rating || 0) || Number(b.id || 0) - Number(a.id || 0))
    .slice(0, limit)
    .map((item) => normalizeRecommendationFromItem(item, t("app.localRecommendationReason")));
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
  catalogSections.innerHTML =
    createRecommendationSection() + availableTypes.map((type) => createSection(type)).join("");
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

async function loadRecommendations(token) {
  try {
    const response = await fetchJson("/api/recommendations/for-you?limit=20", {}, token);
    recommendations = Array.isArray(response?.data) ? response.data : [];
  } catch (error) {
    console.error("Failed to load API recommendations", error);
    recommendations = [];
  }

  if (!recommendations.length) {
    recommendations = buildLocalRecommendationFallback(20);
  }
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

async function toggleWatchLater(btn) {
  const itemId = btn.getAttribute("data-item-id") || "";
  if (!itemId || !currentToken) return;

  const isActive = btn.classList.contains("active");
  btn.disabled = true;

  try {
    await fetchJson(`/watch-later/${itemId}`, { method: isActive ? "DELETE" : "POST" }, currentToken);
    btn.classList.toggle("active", !isActive);
    if (!isActive) btn.classList.add("added");
    else btn.classList.remove("added");
  } catch (error) {
    console.error("Failed to toggle watch later", error);
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

function aiTypeLabel(type) {
  const map = {
    movie: t("app.typeMovie"),
    tv_show: t("app.typeTvShow"),
    game: t("app.typeGame"),
    book: t("app.typeBook"),
  };
  return map[type] || type || "";
}

function aiLaterLabel(type) {
  if (type === "book") return t("app.readLater");
  if (type === "game") return t("app.playLater");
  return t("app.watchLater");
}

function recommendationCardMarkup(rec) {
  const safeTitle = escapeHtml(rec.title || "");
  const safeType = escapeHtml(aiTypeLabel(rec.type));
  const safeDesc = escapeHtml(rec.description || "");
  const safeReason = escapeHtml(rec.reason || "");
  const safeImg = escapeHtml(rec.image || getFallbackImage(rec.title));
  const rating = rec.rating ? `⭐ ${rec.rating}` : "";

  const isDb = rec.is_available_in_app === true && rec.source === "database";
  const statusBadge = isDb
    ? `<span class="rec-badge rec-badge-db">${escapeHtml(t("app.aiAvailableInApp"))}</span>`
    : `<span class="rec-badge rec-badge-ext">${escapeHtml(t("app.aiExternal"))}</span>`;

  let actions = "";
  if (isDb && rec.item_id) {
    const itemId = escapeHtml(String(rec.item_id));
    actions = `
      <div class="rec-actions">
        <button class="rec-action rec-fav" type="button" data-ai-fav data-item-id="${itemId}">❤ ${escapeHtml(t("app.addToFavorites"))}</button>
        <button class="rec-action rec-later" type="button" data-ai-later data-item-id="${itemId}">🕒 ${escapeHtml(aiLaterLabel(rec.type))}</button>
      </div>`;
  } else {
    actions = `<p class="rec-unavailable">${escapeHtml(t("app.aiNotAvailableInApp"))}</p>`;
  }

  const detailAttr = isDb && rec.item_id
    ? `data-detail-url="${escapeHtml(getDetailHref(rec.item_id))}"`
    : "";

  return `
    <article class="rec-card${isDb ? "" : " rec-card-external"}" ${detailAttr}>
      <div class="rec-card-top">
        <img class="rec-card-img" src="${safeImg}" alt="${safeTitle}" loading="lazy" />
        <div class="rec-card-meta">
          <div class="rec-card-head">
            <span class="rec-card-type">${safeType}</span>
            ${rating ? `<span class="rec-card-rating">${rating}</span>` : ""}
          </div>
          <h4 class="rec-card-title">${safeTitle}</h4>
          ${statusBadge}
        </div>
      </div>
      ${safeDesc ? `<p class="rec-card-desc">${safeDesc}</p>` : ""}
      <p class="rec-card-reason">💬 ${safeReason}</p>
      ${actions}
    </article>
  `;
}

function renderRecommendationsInto(botEl, recommendations) {
  if (!Array.isArray(recommendations) || !recommendations.length || !botEl) return;
  const wrap = document.createElement("div");
  wrap.className = "rec-cards";
  wrap.innerHTML = recommendations.map(recommendationCardMarkup).join("");
  botEl.appendChild(wrap);
}

async function requestAIRecommendations(message) {
  const response = await fetchJson(
    "/api/ai-assistant/recommend",
    {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        user_message: message,
        limit: 8,
      }),
    },
    currentToken,
  );

  const data = response?.data || {};
  return {
    reply: String(data.reply || "").trim(),
    recommendations: Array.isArray(data.recommendations) ? data.recommendations : [],
  };
}

function updateSearch(value) {
  searchQuery = value;
  renderCatalog();
  initRowArrows();
}

function handleLanguageChange() {
  renderBanners();

  if (catalogStatusKey) {
    setCatalogStatus(catalogStatusKey);
    return;
  }

  renderCatalog();
  initRowArrows();
}

function openCardDetail(cardEl) {
  const detailUrl = cardEl?.getAttribute("data-detail-url");
  if (!detailUrl) return;
  window.location.href = detailUrl;
}

function initRowArrows() {
  document.querySelectorAll(".row-wrapper").forEach((wrapper) => {
    const row = wrapper.querySelector(".row");
    const leftArrow = wrapper.querySelector(".row-arrow-left");
    const rightArrow = wrapper.querySelector(".row-arrow-right");
    if (!row || !leftArrow || !rightArrow) return;

    const scrollAmount = 440;

    function updateArrows() {
      const atStart = row.scrollLeft <= 2;
      const atEnd = row.scrollLeft + row.clientWidth >= row.scrollWidth - 2;
      leftArrow.classList.toggle("visible", !atStart);
      rightArrow.classList.toggle("visible", !atEnd);
    }

    leftArrow.addEventListener("click", () => {
      row.scrollBy({ left: -scrollAmount, behavior: "smooth" });
    });

    rightArrow.addEventListener("click", () => {
      row.scrollBy({ left: scrollAmount, behavior: "smooth" });
    });

    row.addEventListener("scroll", updateArrows);
    new MutationObserver(updateArrows).observe(row, { childList: true, subtree: true });
    window.addEventListener("resize", updateArrows);

    updateArrows();
  });
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
  } catch (error) {
    console.error("Failed to load favorites", error);
  }

  try {
    await loadRecommendations(token);
  } catch (error) {
    console.error("Failed to load recommendations", error);
  }

  renderCatalog();
  initRowArrows();

  document.getElementById("menuBtn")?.addEventListener("click", toggleSidebar);
  document.getElementById("overlay")?.addEventListener("click", closeSidebar);
  document.getElementById("logoutBtn")?.addEventListener("click", logout);

  document.getElementById("searchInput")?.addEventListener("input", (e) => updateSearch(e.target.value || ""));

  document.body.addEventListener("click", async (e) => {
    const recTypeChip = e.target.closest?.("[data-rec-type]");
    if (recTypeChip) {
      activeRecommendationType = recTypeChip.getAttribute("data-rec-type") || "all";
      renderCatalog();
      initRowArrows();
      return;
    }

    const chip = e.target.closest?.(".chip");
    if (chip) {
      const type = chip.getAttribute("data-type") || "";
      const categoryId = chip.getAttribute("data-category-id") || "all";
      if (type) {
        activeCategoryByType[type] = categoryId;
        renderCatalog();
        initRowArrows();
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

    const aiFav = e.target.closest?.("[data-ai-fav]");
    if (aiFav) {
      e.preventDefault();
      e.stopPropagation();
      await toggleFavorite(aiFav);
      return;
    }

    const aiLater = e.target.closest?.("[data-ai-later]");
    if (aiLater) {
      e.preventDefault();
      e.stopPropagation();
      await toggleWatchLater(aiLater);
      return;
    }

    const navLink = e.target.closest?.("[data-nav='watch-later']");
    if (navLink) {
      e.preventDefault();
      window.location.href = "/pages/watch-later.html";
      return;
    }

    const cardEl = e.target.closest?.(".card-item[data-detail-url]");
    if (cardEl) {
      openCardDetail(cardEl);
      return;
    }

    const aiCardEl = e.target.closest?.(".rec-card[data-detail-url]");
    if (aiCardEl) {
      openCardDetail(aiCardEl);
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
      const result = await requestAIRecommendations(text);
      loadingMsg.remove();

      const safeReply = result.reply || t("app.chatError");
      addMsg("bot", formatChatMessage(safeReply));

      const logs2 = document.getElementById("chatLogs");
      const lastBot = logs2 ? logs2.querySelector(".bot-msg:last-child") : null;
      renderRecommendationsInto(lastBot, result.recommendations);
      if (logs2) logs2.scrollTop = logs2.scrollHeight;

      pushChatHistory("user", text);
      pushChatHistory("assistant", safeReply);
    } catch (error) {
      console.error("Failed to fetch AI recommendations", error);
      loadingMsg.remove();
      addMsg("bot", escapeHtml(t("app.chatError")));
    } finally {
      setChatPendingState(false);
      input?.focus();
    }
  });

  window.addEventListener("trendflix:languagechange", handleLanguageChange);
});
