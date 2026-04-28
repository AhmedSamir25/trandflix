let banners = [];
let editingBannerId = null;

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
        <article class="category-row" data-banner-id="${b.id}">
          <span class="banner-row-info">
            <span class="banner-thumb-wrap">
              <img class="banner-thumb" src="${escapeHtml(b.image_url)}" alt="" loading="lazy" />
            </span>
            <span>
              <strong>${escapeHtml(b.title)}</strong>
              ${b.subtitle ? `<small>${escapeHtml(b.subtitle)}</small>` : ""}
              <code>${b.is_active ? t("admin.bannerActive") : t("admin.bannerInactive")} · #${b.sort_order}</code>
            </span>
          </span>
          <div class="row-actions">
            <button class="text-btn compact-action" type="button" data-edit-banner="${b.id}">${escapeHtml(t("admin.editCategory"))}</button>
            <button class="text-btn danger-btn compact-action" type="button" data-delete-banner="${b.id}">${escapeHtml(t("admin.deleteCategory"))}</button>
          </div>
        </article>`,
    )
    .join("");
}

function resetBannerForm() {
  document.getElementById("bannerForm")?.reset();
  document.getElementById("bannerIsActive").checked = true;
  editingBannerId = null;
  document.getElementById("bannerFormTitle")?.setAttribute("data-i18n", "admin.bannerCreateTitle");
  document.getElementById("bannerFormTitle").textContent = t("admin.bannerCreateTitle");
  setButtonLoading("saveBannerBtn", "admin.bannerCreating", "admin.bannerCreate", false);
  document.getElementById("cancelEditBtn").hidden = true;
}

function startBannerEdit(bannerId) {
  const banner = banners.find((b) => Number(b.id) === Number(bannerId));
  if (!banner) return;

  editingBannerId = banner.id;

  const titleInput = document.getElementById("bannerTitle");
  const subtitleInput = document.getElementById("bannerSubtitle");
  const imageInput = document.getElementById("bannerImageUrl");
  const linkInput = document.getElementById("bannerLinkUrl");
  const sortInput = document.getElementById("bannerSortOrder");
  const activeInput = document.getElementById("bannerIsActive");
  const formTitle = document.getElementById("bannerFormTitle");
  const saveBtn = document.getElementById("saveBannerBtn");
  const cancelBtn = document.getElementById("cancelEditBtn");

  if (titleInput) titleInput.value = banner.title || "";
  if (subtitleInput) subtitleInput.value = banner.subtitle || "";
  if (imageInput) imageInput.value = banner.image_url || "";
  if (linkInput) linkInput.value = banner.link_url || "";
  if (sortInput) sortInput.value = banner.sort_order ?? 0;
  if (activeInput) activeInput.checked = Boolean(banner.is_active);
  if (formTitle) formTitle.textContent = t("admin.bannerEditTitle");
  if (saveBtn) saveBtn.textContent = t("admin.bannerSave");
  if (cancelBtn) cancelBtn.hidden = false;

  document.getElementById("bannerForm")?.scrollIntoView({ behavior: "smooth", block: "start" });
}

async function loadBanners() {
  clearNotice("pageError");
  const data = await fetchJson("/banners/all", { headers: authHeaders() });
  banners = Array.isArray(data?.banners) ? data.banners : [];
  renderBannerList();
}

async function handleBannerSubmit(event) {
  event.preventDefault();
  clearNotice("pageError");
  clearNotice("bannerStatus");

  const title = String(document.getElementById("bannerTitle")?.value || "").trim();
  const subtitle = String(document.getElementById("bannerSubtitle")?.value || "").trim();
  const imageUrl = String(document.getElementById("bannerImageUrl")?.value || "").trim();
  const linkUrl = String(document.getElementById("bannerLinkUrl")?.value || "").trim();
  const sortOrder = Number(document.getElementById("bannerSortOrder")?.value ?? 0);
  const isActive = Boolean(document.getElementById("bannerIsActive")?.checked);

  if (!title) {
    setNotice("bannerStatus", t("admin.bannerTitleRequired"), "error");
    return;
  }
  if (!imageUrl) {
    setNotice("bannerStatus", t("admin.bannerImageRequired"), "error");
    return;
  }

  const isEditing = Boolean(editingBannerId);
  setButtonLoading(
    "saveBannerBtn",
    isEditing ? "admin.bannerSaving" : "admin.bannerCreating",
    isEditing ? "admin.bannerSave" : "admin.bannerCreate",
    true,
  );

  try {
    const response = await fetchJson(
      isEditing ? `/banners/${editingBannerId}` : "/banners/",
      {
        method: isEditing ? "PUT" : "POST",
        headers: authHeaders({ "Content-Type": "application/json" }),
        body: JSON.stringify({ title, subtitle, image_url: imageUrl, link_url: linkUrl, sort_order: sortOrder, is_active: isActive }),
      },
    );

    resetBannerForm();
    setNotice("bannerStatus", response?.msg || t(isEditing ? "admin.bannerUpdated" : "admin.bannerCreated"), "success");
    await loadBanners();
  } catch (error) {
    setNotice("bannerStatus", error.message, "error");
  } finally {
    const stillEditing = Boolean(editingBannerId);
    setButtonLoading(
      "saveBannerBtn",
      stillEditing ? "admin.bannerSaving" : "admin.bannerCreating",
      stillEditing ? "admin.bannerSave" : "admin.bannerCreate",
      false,
    );
  }
}

async function deleteBanner(bannerId) {
  const banner = banners.find((b) => Number(b.id) === Number(bannerId));
  const name = banner?.title || t("admin.thisBanner");
  if (!window.confirm(`Delete banner "${name}"?`)) return;

  clearNotice("pageError");
  clearNotice("bannerStatus");

  const response = await fetchJson(`/banners/${bannerId}`, {
    method: "DELETE",
    headers: authHeaders(),
  });

  if (Number(editingBannerId) === Number(bannerId)) resetBannerForm();
  setNotice("bannerStatus", response?.msg || t("admin.bannerDeleted"), "success");
  await loadBanners();
}

window.addEventListener("DOMContentLoaded", async () => {
  if (!requireAdmin()) return;

  bindLogout();
  highlightActiveNav();

  document.getElementById("bannerForm")?.addEventListener("submit", handleBannerSubmit);

  document.getElementById("refreshBannersBtn")?.addEventListener("click", () =>
    loadBanners().catch(showPageError),
  );

  document.getElementById("cancelEditBtn")?.addEventListener("click", resetBannerForm);

  document.getElementById("bannerList")?.addEventListener("click", (event) => {
    const editBtn = event.target.closest("[data-edit-banner]");
    if (editBtn) {
      startBannerEdit(editBtn.dataset.editBanner);
      return;
    }

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
