(() => {
  const token = localStorage.getItem("trendflix.token");
  if (token) {
    window.location.replace("/pages/app.html");
    return;
  }

  window.addEventListener("DOMContentLoaded", () => {
    const form = document.getElementById("getStartedForm");
    form.addEventListener("submit", (e) => {
      e.preventDefault();
      const fd = new FormData(form);
      const email = String(fd.get("email") || "").trim();
      if (!email) return;
      window.location.href = `/pages/auth/auth.html?email=${encodeURIComponent(email)}`;
    });
  });
})();
