let categories = [];
let selectedCategoryIds = [];

function getFormMode() {
  return document.body.dataset.itemFormMode === "edit" ? "edit" : "create";
}

function getItemIdFromUrl() {
  const rawId = new URLSearchParams(window.location.search).get("id");
  const id = Number.parseInt(String(rawId || "").trim(), 10);
  return Number.isInteger(id) && id > 0 ? id : null;
}

function getSelectedCategoryIds() {
  const categoryInputs = document.querySelectorAll('[name="category_ids"]');
  const selectedFromDom = Array.from(document.querySelectorAll('[name="category_ids"]:checked'))
    .map((input) => Number(input.value))
    .filter((id) => Number.isInteger(id) && id > 0);

  return categoryInputs.length ? selectedFromDom : selectedCategoryIds;
}

function renderItemCategoryOptions() {
  const container = document.getElementById("itemCategoryList");
  if (!container) return;

  if (!categories.length) {
    container.innerHTML = `<p class="notice">${escapeHtml(t("admin.noCategoriesCreate"))}</p>`;
    return;
  }

  const activeIds = new Set(getSelectedCategoryIds());
  container.innerHTML = categories
    .map((category) => {
      const checked = activeIds.has(category.id) ? "checked" : "";
      return `
        <label class="checkbox-card">
          <input type="checkbox" name="category_ids" value="${category.id}" ${checked} />
          <span>
            <strong>${escapeHtml(category.name)}</strong>
            <small>${escapeHtml(category.slug)}</small>
          </span>
        </label>
      `;
    })
    .join("");
}

function syncTypeFields() {
  const type = document.getElementById("itemType")?.value || "movie";
  document.querySelectorAll("[data-field-for]").forEach((field) => {
    const visible = field.getAttribute("data-field-for") === type;
    field.classList.toggle("is-hidden", !visible);
    field.querySelectorAll("input, select, textarea").forEach((input) => {
      input.disabled = !visible;
    });
  });
}

function updateImagePreview(src) {
  const block = document.getElementById("imagePreviewBlock");
  const preview = document.getElementById("imagePreview");
  if (!block || !preview) return;

  if (!src) {
    block.hidden = true;
    preview.removeAttribute("src");
    return;
  }

  preview.src = src;
  block.hidden = false;
}

async function uploadCoverImageIfNeeded() {
  const fileInput = document.getElementById("coverImageFile");
  const coverInput = document.getElementById("coverImageInput");
  const file = fileInput?.files?.[0];

  if (!file) return String(coverInput?.value || "").trim();

  setNotice("itemStatus", t("admin.compressingImage"));
  const uploadFile = await compressImageFile(file, 0.75);

  setNotice("itemStatus", t("admin.uploadingImage"));
  const formData = new FormData();
  formData.append("file", uploadFile);

  const response = await fetchJson("/upload/item-image", {
    method: "POST",
    headers: authHeaders(),
    body: formData,
  });

  const path = String(response?.path || "").trim();
  if (coverInput) coverInput.value = path;
  updateImagePreview(path);
  return path;
}

function parseOptionalInteger(value) {
  const parsed = Number.parseInt(String(value || "").trim(), 10);
  return Number.isInteger(parsed) && parsed > 0 ? parsed : null;
}

function buildItemPayload(form) {
  const formData = new FormData(form);
  const type = String(formData.get("type") || "").trim();
  const payload = {
    title: String(formData.get("title") || "").trim(),
    description: String(formData.get("description") || "").trim(),
    type,
    cover_image: String(formData.get("cover_image") || "").trim(),
    content_link: String(formData.get("content_link") || "").trim() || null,
    release_date: String(formData.get("release_date") || "").trim(),
    rating: Number(formData.get("rating") || 0),
    category_ids: getSelectedCategoryIds(),
  };

  if (!payload.title || !payload.type || !payload.release_date) {
    throw new Error(t("admin.itemTitleRequired"));
  }

  if (type === "book") {
    payload.author = String(formData.get("author") || "").trim() || null;
    payload.pages_count = parseOptionalInteger(formData.get("pages_count"));
  }
  if (type === "movie") {
    payload.director = String(formData.get("director") || "").trim() || null;
    payload.duration = parseOptionalInteger(formData.get("duration"));
  }
  if (type === "game") {
    payload.developer = String(formData.get("developer") || "").trim() || null;
    payload.platform = String(formData.get("platform") || "").trim() || null;
  }

  return payload;
}

function formatDateForInput(value) {
  return String(value || "").split("T")[0].trim();
}

function setFieldValue(name, value) {
  document.querySelectorAll(`[name="${name}"]`).forEach((input) => {
    input.value = value == null ? "" : String(value);
  });
}

function populateItemForm(item) {
  const form = document.getElementById("itemForm");
  if (!form) return;

  setFieldValue("title", item.title);
  setFieldValue("description", item.description);
  setFieldValue("type", item.type || "movie");
  syncTypeFields();

  setFieldValue("cover_image", item.cover_image);
  setFieldValue("content_link", item.content_link);
  setFieldValue("release_date", formatDateForInput(item.release_date));
  setFieldValue("rating", item.rating || 0);
  setFieldValue("author", item.author);
  setFieldValue("director", item.director);
  setFieldValue("developer", item.developer);
  setFieldValue("duration", item.duration);
  setFieldValue("pages_count", item.pages_count);
  setFieldValue("platform", item.platform);

  selectedCategoryIds = Array.isArray(item.categories)
    ? item.categories
        .map((category) => Number(category.id))
        .filter((id) => Number.isInteger(id) && id > 0)
    : [];

  renderItemCategoryOptions();
  updateImagePreview(String(item.cover_image || "").trim());
}

async function loadCategories() {
  const data = await fetchJson("/categories");
  categories = Array.isArray(data?.categories) ? data.categories : [];
  renderItemCategoryOptions();
}

async function loadItemForEdit() {
  const itemId = getItemIdFromUrl();
  if (!itemId) {
    throw new Error(t("admin.invalidItemId"));
  }

  const data = await fetchJson(`/items/${itemId}`);
  if (!data?.item) {
    throw new Error(t("admin.itemLoadFailed"));
  }

  populateItemForm(data.item);
}

async function handleItemSubmit(event) {
  event.preventDefault();
  clearNotice("pageError");
  clearNotice("itemStatus");

  const mode = getFormMode();
  const itemId = getItemIdFromUrl();
  const isEditMode = mode === "edit";
  const buttonIdleKey = isEditMode ? "admin.saveChanges" : "admin.createItem";
  const buttonLoadingKey = isEditMode ? "admin.savingChanges" : "admin.creatingItem";

  setButtonLoading("itemSubmitBtn", buttonLoadingKey, buttonIdleKey, true);

  try {
    if (isEditMode && !itemId) {
      throw new Error(t("admin.invalidItemId"));
    }

    await uploadCoverImageIfNeeded();
    const payload = buildItemPayload(event.target);

    const response = await fetchJson(isEditMode ? `/items/${itemId}` : "/items", {
      method: isEditMode ? "PUT" : "POST",
      headers: authHeaders({ "Content-Type": "application/json" }),
      body: JSON.stringify(payload),
    });

    if (isEditMode) {
      if (response?.item) {
        populateItemForm(response.item);
      }
      setNotice("itemStatus", response?.msg || t("admin.itemUpdated"), "success");
      return;
    }

    event.target.reset();
    selectedCategoryIds = [];
    updateImagePreview("");
    syncTypeFields();
    renderItemCategoryOptions();
    setNotice("itemStatus", response?.msg || t("admin.itemCreated"), "success");
  } catch (error) {
    setNotice("itemStatus", error.message, "error");
  } finally {
    setButtonLoading("itemSubmitBtn", buttonLoadingKey, buttonIdleKey, false);
  }
}

window.addEventListener("DOMContentLoaded", async () => {
  if (!requireAdmin()) return;

  bindLogout();
  highlightActiveNav();
  syncTypeFields();

  document.getElementById("itemForm")?.addEventListener("submit", handleItemSubmit);
  document.getElementById("itemType")?.addEventListener("change", syncTypeFields);

  document.getElementById("coverImageInput")?.addEventListener("input", (event) => {
    const fileInput = document.getElementById("coverImageFile");
    if (fileInput) fileInput.value = "";
    updateImagePreview(String(event.target.value || "").trim());
  });

  document.getElementById("coverImageFile")?.addEventListener("change", (event) => {
    const file = event.target.files?.[0];
    if (!file) {
      updateImagePreview(String(document.getElementById("coverImageInput")?.value || "").trim());
      return;
    }

    updateImagePreview(URL.createObjectURL(file));
  });

  window.addEventListener("trendflix:languagechange", renderItemCategoryOptions);

  try {
    await loadCategories();
    if (getFormMode() === "edit") {
      await loadItemForEdit();
    }
  } catch (error) {
    showPageError(error);
  }
});
