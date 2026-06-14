let editingBannerId = null;

function getFormMode() {
  return document.body.dataset.bannerFormMode === "edit" ? "edit" : "create";
}

function getBannerIdFromUrl() {
  const rawId = new URLSearchParams(window.location.search).get("id");
  const id = Number.parseInt(String(rawId || "").trim(), 10);
  return Number.isInteger(id) && id > 0 ? id : null;
}

function resetBannerForm() {
  document.getElementById("bannerForm")?.reset();
  const activeInput = document.getElementById("bannerIsActive");
  if (activeInput) activeInput.checked = true;
  setButtonLoading("saveBannerBtn", "admin.bannerCreating", "admin.bannerCreate", false);
}

function populateBannerForm(banner) {
  const titleInput = document.getElementById("bannerTitle");
  const subtitleInput = document.getElementById("bannerSubtitle");
  const titleArInput = document.getElementById("bannerTitleAr");
  const subtitleArInput = document.getElementById("bannerSubtitleAr");
  const imageInput = document.getElementById("bannerImageUrl");
  const linkInput = document.getElementById("bannerLinkUrl");
  const sortInput = document.getElementById("bannerSortOrder");
  const activeInput = document.getElementById("bannerIsActive");

  if (titleInput) titleInput.value = banner.title || "";
  if (subtitleInput) subtitleInput.value = banner.subtitle || "";
  if (titleArInput) titleArInput.value = banner.title_ar || "";
  if (subtitleArInput) subtitleArInput.value = banner.subtitle_ar || "";
  if (imageInput) imageInput.value = banner.image_url || "";
  if (linkInput) linkInput.value = banner.link_url || "";
  if (sortInput) sortInput.value = banner.sort_order ?? 0;
  if (activeInput) activeInput.checked = Boolean(banner.is_active);
}

async function loadBannerForEdit() {
  const bannerId = getBannerIdFromUrl();
  if (!bannerId) {
    throw new Error(t("admin.invalidBannerId"));
  }

  editingBannerId = bannerId;

  const data = await fetchJson("/banners/all", { headers: authHeaders() });
  const allBanners = Array.isArray(data?.banners) ? data.banners : [];
  const banner = allBanners.find((entry) => Number(entry.id) === Number(bannerId));

  if (!banner) {
    throw new Error(t("admin.bannerLoadFailed"));
  }

  populateBannerForm(banner);
}

async function handleBannerSubmit(event) {
  event.preventDefault();
  clearNotice("pageError");
  clearNotice("bannerStatus");

  const title = String(document.getElementById("bannerTitle")?.value || "").trim();
  const subtitle = String(document.getElementById("bannerSubtitle")?.value || "").trim();
  const titleAr = String(document.getElementById("bannerTitleAr")?.value || "").trim();
  const subtitleAr = String(document.getElementById("bannerSubtitleAr")?.value || "").trim();
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

  const isEditing = getFormMode() === "edit";
  const buttonIdleKey = isEditing ? "admin.bannerSave" : "admin.bannerCreate";
  const buttonLoadingKey = isEditing ? "admin.bannerSaving" : "admin.bannerCreating";

  setButtonLoading("saveBannerBtn", buttonLoadingKey, buttonIdleKey, true);

  try {
    if (isEditing && !editingBannerId) {
      throw new Error(t("admin.invalidBannerId"));
    }

    const response = await fetchJson(isEditing ? `/banners/${editingBannerId}` : "/banners/", {
      method: isEditing ? "PUT" : "POST",
      headers: authHeaders({ "Content-Type": "application/json" }),
      body: JSON.stringify({
        title,
        subtitle,
        title_ar: titleAr,
        subtitle_ar: subtitleAr,
        image_url: imageUrl,
        link_url: linkUrl,
        sort_order: sortOrder,
        is_active: isActive,
      }),
    });

    if (isEditing) {
      setNotice("bannerStatus", response?.msg || t("admin.bannerUpdated"), "success");
      return;
    }

    resetBannerForm();
    setNotice("bannerStatus", response?.msg || t("admin.bannerCreated"), "success");
  } catch (error) {
    setNotice("bannerStatus", error.message, "error");
  } finally {
    setButtonLoading("saveBannerBtn", buttonLoadingKey, buttonIdleKey, false);
  }
}

window.addEventListener("DOMContentLoaded", async () => {
  if (!requireAdmin()) return;

  bindLogout();
  highlightActiveNav();

  document.getElementById("bannerForm")?.addEventListener("submit", handleBannerSubmit);

  try {
    if (getFormMode() === "edit") {
      await loadBannerForEdit();
    }
  } catch (error) {
    showPageError(error);
  }
});
