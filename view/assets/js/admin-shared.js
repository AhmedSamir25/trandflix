const TOKEN_KEY = "trendflix.token";
const FALLBACK_IMAGE_BASE = "https://placehold.co/500x700/0f172a/f8fafc";

function t(key) {
  return window.TrendFlixI18n?.t(key) ?? key;
}

function getToken() {
  return localStorage.getItem(TOKEN_KEY);
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

function requireAdmin() {
  const token = getToken();
  if (!token) {
    window.location.replace("/pages/auth/auth.html");
    return false;
  }
  const payload = parseJwtPayload(token);
  if (String(payload?.role || "").trim().toLowerCase() !== "admin") {
    window.location.replace("/pages/app.html");
    return false;
  }
  return true;
}

function authHeaders(headers = {}) {
  return { ...headers, Authorization: `Bearer ${getToken()}` };
}

function escapeHtml(value) {
  return String(value)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#039;");
}

function slugify(value) {
  return String(value)
    .trim()
    .toLowerCase()
    .replace(/[^\p{L}\p{N}]+/gu, "-")
    .replace(/^-+|-+$/g, "");
}

function setNotice(elementId, message, tone = "info") {
  const el = document.getElementById(elementId);
  if (!el) return;
  el.hidden = !message;
  el.textContent = message;
  el.className = "notice";
  if (tone === "success") el.classList.add("success-notice");
  if (tone === "error") el.classList.add("error-notice");
}

function clearNotice(elementId) {
  setNotice(elementId, "");
}

function setButtonLoading(buttonId, loadingKey, idleKey, isLoading) {
  const btn = document.getElementById(buttonId);
  if (!btn) return;
  btn.disabled = isLoading;
  btn.textContent = t(isLoading ? loadingKey : idleKey);
}

function getFallbackImage(title) {
  return `${FALLBACK_IMAGE_BASE}?text=${encodeURIComponent(title || "TrendFlix")}`;
}

function loadImageFromFile(file) {
  return new Promise((resolve, reject) => {
    const image = new Image();
    const objectUrl = URL.createObjectURL(file);

    image.onload = () => {
      URL.revokeObjectURL(objectUrl);
      resolve(image);
    };

    image.onerror = () => {
      URL.revokeObjectURL(objectUrl);
      reject(new Error("Unable to read image file"));
    };

    image.src = objectUrl;
  });
}

async function compressImageFile(file, quality = 0.75) {
  if (!(file instanceof File) || !String(file.type || "").startsWith("image/")) {
    return file;
  }

  if (file.type === "image/gif") {
    return file;
  }

  const image = await loadImageFromFile(file);
  const canvas = document.createElement("canvas");
  canvas.width = image.naturalWidth || image.width;
  canvas.height = image.naturalHeight || image.height;

  const context = canvas.getContext("2d");
  if (!context) {
    return file;
  }

  context.drawImage(image, 0, 0, canvas.width, canvas.height);

  const blob = await new Promise((resolve) => {
    canvas.toBlob(resolve, "image/webp", quality);
  });

  if (!blob || blob.size >= file.size) {
    return file;
  }

  const baseName = file.name.replace(/\.[^.]+$/, "") || "image";
  return new File([blob], `${baseName}.webp`, {
    type: "image/webp",
    lastModified: Date.now(),
  });
}

async function fetchJson(url, options = {}) {
  const response = await fetch(url, options);
  const data = await response.json().catch(() => ({}));
  if (!response.ok) throw new Error(data?.msg || `Request failed: ${response.status}`);
  return data;
}

function showPageError(error) {
  const message = error?.message || String(error || "");
  setNotice("pageError", message, "error");
}

function bindLogout() {
  document.getElementById("logoutBtn")?.addEventListener("click", () => {
    localStorage.removeItem(TOKEN_KEY);
    window.location.replace("/pages/auth/auth.html");
  });
}

function highlightActiveNav() {
  const path = window.location.pathname;
  document.querySelectorAll(".admin-nav-link").forEach((link) => {
    link.classList.toggle("active", link.getAttribute("href") === path);
  });
}
