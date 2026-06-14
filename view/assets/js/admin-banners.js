let banners = [];

function renderBannerList() {
  const list = document.getElementById("bannerList");
  if (!list) return;

  if (!banners.length) {
    list.innerHTML = `<p class="notice">${escapeHtml(t("admin.noBanners"))}</p>`;
    return;
  }

  list.innerHTML = banners
    .map(
      (b) => `
        <article class="category-row banner-card" data-banner-id="${b.id}">
          <span class="banner-thumb-wrap">
            <img class="banner-thumb" src="${escapeHtml(b.image_url)}" alt="" loading="lazy" />
          </span>
          <div class="banner-card-body">
            <strong class="banner-card-title">${escapeHtml(window.TrendFlixI18n?.localizedText?.(b.title, b.title_ar) || b.title || b.title_ar || "")}</strong>
            ${window.TrendFlixI18n?.localizedText?.(b.subtitle, b.subtitle_ar) ? `<small class="banner-card-subtitle">${escapeHtml(window.TrendFlixI18n.localizedText(b.subtitle, b.subtitle_ar))}</small>` : ""}
          </div>
          <div class="row-actions">
            <a class="text-btn compact-action" href="/pages/admin/edit-banner.html?id=${b.id}">${escapeHtml(t("admin.editCategory"))}</a>
            <button class="text-btn danger-btn compact-action" type="button" data-delete-banner="${b.id}">${escapeHtml(t("admin.deleteCategory"))}</button>
          </div>
        </article>`,
    )
    .join("");
}

async function loadBanners() {
  clearNotice("pageError");
  const data = await fetchJson("/banners/all", { headers: authHeaders() });
  banners = Array.isArray(data?.banners) ? data.banners : [];
  renderBannerList();
}

async function deleteBanner(bannerId) {
  const banner = banners.find((b) => Number(b.id) === Number(bannerId));
  const name = banner?.title || t("admin.thisBanner");
  if (!window.confirm(`Delete banner "${name}"?`)) return;

  clearNotice("pageError");

  const response = await fetchJson(`/banners/${bannerId}`, {
    method: "DELETE",
    headers: authHeaders(),
  });

  setNotice("pageError", response?.msg || t("admin.bannerDeleted"), "success");
  await loadBanners();
}

window.addEventListener("DOMContentLoaded", async () => {
  if (!requireAdmin()) return;

  bindLogout();
  highlightActiveNav();

  document.getElementById("bannerList")?.addEventListener("click", (event) => {
    const deleteBtn = event.target.closest("[data-delete-banner]");
    if (deleteBtn) {
      deleteBanner(deleteBtn.dataset.deleteBanner).catch(showPageError);
    }
  });

  window.addEventListener("trendflix:languagechange", renderBannerList);

  try {
    await loadBanners();
  } catch (error) {
    showPageError(error);
  }
});
