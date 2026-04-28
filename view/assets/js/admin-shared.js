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

function isTokenExpired(payload) {
  const exp = Number(payload?.exp) || 0;
  return exp > 0 && Date.now() >= exp * 1000;
}

function redirectToLogin() {
  localStorage.removeItem(TOKEN_KEY);
  window.location.replace("/pages/auth/auth.html");
}

function requireAdmin() {
  const token = getToken();
  if (!token) {
    redirectToLogin();
    return false;
  }
  const payload = parseJwtPayload(token);
  if (!payload || isTokenExpired(payload)) {
    redirectToLogin();
    return false;
  }
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
  if (!response.ok) {
    const error = new Error(data?.msg || `Request failed: ${response.status}`);
    error.status = response.status;
    error.data = data;
    throw error;
  }
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
  document.querySelectorAll(".admin-nav-link, .admin-drawer-link").forEach((link) => {
    const href = link.getAttribute("href");
    if (!href) return;
    link.classList.toggle("active", href === path);
  });
}

/* ─── Admin drawer (slide-in sidebar) ─────────────────────── */

const ADMIN_DRAWER_LINKS = [
  { href: "/pages/admin.html",                icon: "🏠", labelKey: "admin.navDashboard",   fallback: "Dashboard" },
  { href: "/pages/admin/categories.html",     icon: "🗂️", labelKey: "admin.navCategories",  fallback: "Categories" },
  { href: "/pages/admin/create-item.html",    icon: "➕", labelKey: "admin.navCreateItem",  fallback: "Create Item" },
  { href: "/pages/admin/catalog.html",        icon: "📋", labelKey: "admin.navCatalog",     fallback: "All Items" },
];

function injectAdminDrawer() {
  if (document.getElementById("adminDrawer")) return;

  const linksHtml = ADMIN_DRAWER_LINKS.map(
    (link) => `
      <a class="admin-drawer-link" href="${link.href}">
        <span class="admin-drawer-icon">${link.icon}</span>
        <span data-i18n="${link.labelKey}">${escapeHtml(link.fallback)}</span>
      </a>
    `,
  ).join("");

  const drawer = document.createElement("aside");
  drawer.id = "adminDrawer";
  drawer.className = "admin-drawer";
  drawer.setAttribute("aria-hidden", "true");
  drawer.setAttribute("aria-label", "Admin navigation");
  drawer.innerHTML = `
    <div class="admin-drawer-header">
      <span class="admin-drawer-brand">TrendFlix</span>
      <span class="admin-drawer-tag" data-i18n="admin.eyebrow">Admin Dashboard</span>
    </div>
    <nav class="admin-drawer-nav">
      ${linksHtml}
      <div class="admin-drawer-divider"></div>
      <a class="admin-drawer-link" href="/pages/app.html">
        <span class="admin-drawer-icon">↩</span>
        <span data-i18n="admin.backToApp">Back to app</span>
      </a>
      <button class="admin-drawer-link admin-drawer-danger" type="button" id="adminDrawerLogout">
        <span class="admin-drawer-icon">🚪</span>
        <span data-i18n="admin.logout">Logout</span>
      </button>
    </nav>
  `;

  const overlay = document.createElement("div");
  overlay.id = "adminDrawerOverlay";
  overlay.className = "admin-drawer-overlay";
  overlay.setAttribute("aria-hidden", "true");

  const menuBtn = document.createElement("button");
  menuBtn.id = "adminMenuBtn";
  menuBtn.type = "button";
  menuBtn.className = "admin-menu-btn";
  menuBtn.setAttribute("aria-label", "Toggle admin menu");
  menuBtn.setAttribute("aria-controls", "adminDrawer");
  menuBtn.setAttribute("aria-expanded", "false");
  menuBtn.innerHTML = `<span></span><span></span><span></span>`;

  document.body.appendChild(overlay);
  document.body.appendChild(drawer);
  document.body.appendChild(menuBtn);

  bindAdminDrawer();
  window.TrendFlixI18n?.translatePage?.();
}

function openAdminDrawer() {
  document.getElementById("adminDrawer")?.classList.add("open");
  document.getElementById("adminDrawer")?.setAttribute("aria-hidden", "false");
  document.getElementById("adminDrawerOverlay")?.classList.add("visible");
  const btn = document.getElementById("adminMenuBtn");
  if (btn) {
    btn.classList.add("open");
    btn.setAttribute("aria-expanded", "true");
  }
  document.body.classList.add("admin-drawer-open");
}

function closeAdminDrawer() {
  document.getElementById("adminDrawer")?.classList.remove("open");
  document.getElementById("adminDrawer")?.setAttribute("aria-hidden", "true");
  document.getElementById("adminDrawerOverlay")?.classList.remove("visible");
  const btn = document.getElementById("adminMenuBtn");
  if (btn) {
    btn.classList.remove("open");
    btn.setAttribute("aria-expanded", "false");
  }
  document.body.classList.remove("admin-drawer-open");
}

function toggleAdminDrawer() {
  const drawer = document.getElementById("adminDrawer");
  if (!drawer) return;
  if (drawer.classList.contains("open")) closeAdminDrawer();
  else openAdminDrawer();
}

function bindAdminDrawer() {
  document.getElementById("adminMenuBtn")?.addEventListener("click", toggleAdminDrawer);
  document.getElementById("adminDrawerOverlay")?.addEventListener("click", closeAdminDrawer);

  document.addEventListener("keydown", (event) => {
    if (event.key === "Escape") closeAdminDrawer();
  });

  document.getElementById("adminDrawerLogout")?.addEventListener("click", () => {
    localStorage.removeItem(TOKEN_KEY);
    window.location.replace("/pages/auth/auth.html");
  });
}

window.addEventListener("DOMContentLoaded", () => {
  injectAdminDrawer();
  highlightActiveNav();
});
