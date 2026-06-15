const COMM_TOKEN_KEY = "trendflix.token";

const COMM_FALLBACK_AVATAR =
  "https://placehold.co/120x120/0f172a/f8fafc?text=Community";
const COMM_FALLBACK_COVER =
  "https://placehold.co/1200x320/0f172a/f8fafc?text=TrendFlix";

function commToken() {
  return localStorage.getItem(COMM_TOKEN_KEY) || "";
}

function commAuthHeaders() {
  const token = commToken();
  return token ? { Authorization: `Bearer ${token}` } : {};
}

function commCurrentUserId() {
  const token = commToken();
  if (!token) return 0;
  try {
    const payload = token.split(".")[1] || "";
    const normalized = payload.replaceAll("-", "+").replaceAll("_", "/");
    const padded = normalized.padEnd(Math.ceil(normalized.length / 4) * 4, "=");
    return Number(JSON.parse(window.atob(padded))?.sub) || 0;
  } catch {
    return 0;
  }
}

function commRedirectIfUnauth() {
  if (!commToken()) {
    window.location.replace("/pages/auth/auth.html");
    return true;
  }
  return false;
}

function commEscapeHtml(value) {
  return String(value ?? "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#039;");
}

function commAvatarUrl(community) {
  return community?.avatar_image || COMM_FALLBACK_AVATAR;
}

function commCoverUrl(community) {
  return community?.cover_image || COMM_FALLBACK_COVER;
}

function commT(key) {
  return window.TrendFlixI18n?.t(key) ?? key;
}

function commFormatDate(dateString) {
  if (!dateString) return "";
  const date = new Date(dateString);
  if (isNaN(date.getTime())) return "";
  const lang = window.TrendFlixI18n?.getLang?.() === "ar" ? "ar-EG" : "en-US";
  return date.toLocaleDateString(lang, {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

function commCategoryLabel(type) {
  const map = {
    movies: "communities.catMovies",
    series: "communities.catSeries",
    books: "communities.catBooks",
    games: "communities.catGames",
    mixed: "communities.catMixed",
  };
  const key = map[type];
  if (key) return commT(key);
  return type ? commT(`communities.cat${type.charAt(0).toUpperCase()}${type.slice(1)}`) : commT("communities.catMixed");
}

async function commFetch(path, options = {}) {
  const headers = {
    "Content-Type": "application/json",
    Accept: "application/json",
    ...commAuthHeaders(),
    ...(options.headers || {}),
  };
  const res = await fetch(path, { ...options, headers });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    const message = data?.message || data?.msg || `Request failed (${res.status})`;
    const err = new Error(message);
    err.status = res.status;
    err.payload = data;
    throw err;
  }
  return data;
}
