window.addEventListener("DOMContentLoaded", () => {
  if (!requireAdmin()) return;
  bindLogout();
  highlightActiveNav();
});
