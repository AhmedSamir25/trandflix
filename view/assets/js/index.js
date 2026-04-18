(() => {
  const token = localStorage.getItem("trendflix.token");
  if (token) {
    window.location.replace("/pages/app.html");
    return;
  }
  window.location.replace("/pages/index.html");
})();
