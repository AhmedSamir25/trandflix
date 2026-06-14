let slugTouched = false;
let editingCategoryId = null;

function getFormMode() {
  return document.body.dataset.categoryFormMode === "edit" ? "edit" : "create";
}

function getCategoryIdFromUrl() {
  const rawId = new URLSearchParams(window.location.search).get("id");
  const id = Number.parseInt(String(rawId || "").trim(), 10);
  return Number.isInteger(id) && id > 0 ? id : null;
}

function resetCategoryForm() {
  document.getElementById("categoryForm")?.reset();
  slugTouched = false;
  setButtonLoading("createCategoryBtn", "admin.creatingCategory", "admin.createCategory", false);
}

function populateCategoryForm(category) {
  const nameInput = document.getElementById("categoryName");
  const nameArInput = document.getElementById("categoryNameAr");
  const slugInput = document.getElementById("categorySlug");

  if (nameInput) nameInput.value = category.name || "";
  if (nameArInput) nameArInput.value = category.name_ar || "";
  if (slugInput) slugInput.value = category.slug || "";
}

async function loadCategoryForEdit() {
  const categoryId = getCategoryIdFromUrl();
  if (!categoryId) {
    throw new Error(t("admin.invalidCategoryId"));
  }

  editingCategoryId = categoryId;
  slugTouched = true;

  const data = await fetchJson("/categories");
  const categories = Array.isArray(data?.categories) ? data.categories : [];
  const category = categories.find((entry) => Number(entry.id) === Number(categoryId));

  if (!category) {
    throw new Error(t("admin.categoryLoadFailed"));
  }

  populateCategoryForm(category);
}

async function handleCategorySubmit(event) {
  event.preventDefault();
  clearNotice("pageError");
  clearNotice("categoryStatus");

  const nameInput = document.getElementById("categoryName");
  const nameArInput = document.getElementById("categoryNameAr");
  const slugInput = document.getElementById("categorySlug");
  const name = String(nameInput?.value || "").trim();
  const nameAr = String(nameArInput?.value || "").trim();
  const slug = slugify(slugInput?.value || name);

  if (!name) {
    setNotice("categoryStatus", t("admin.categoryNameRequired"), "error");
    return;
  }

  if (slugInput) slugInput.value = slug;

  const isEditing = getFormMode() === "edit";
  const buttonIdleKey = isEditing ? "admin.saveCategory" : "admin.createCategory";
  const buttonLoadingKey = isEditing ? "admin.savingCategory" : "admin.creatingCategory";

  setButtonLoading("createCategoryBtn", buttonLoadingKey, buttonIdleKey, true);

  try {
    if (isEditing && !editingCategoryId) {
      throw new Error(t("admin.invalidCategoryId"));
    }

    const response = await fetchJson(isEditing ? `/categories/${editingCategoryId}` : "/categories", {
      method: isEditing ? "PUT" : "POST",
      headers: authHeaders({ "Content-Type": "application/json" }),
      body: JSON.stringify({ name, name_ar: nameAr, slug }),
    });

    if (isEditing) {
      setNotice("categoryStatus", response?.msg || t("admin.categoryUpdated"), "success");
      return;
    }

    resetCategoryForm();
    setNotice("categoryStatus", response?.msg || t("admin.categoryCreated"), "success");
  } catch (error) {
    setNotice("categoryStatus", error.message, "error");
  } finally {
    setButtonLoading("createCategoryBtn", buttonLoadingKey, buttonIdleKey, false);
  }
}

window.addEventListener("DOMContentLoaded", async () => {
  if (!requireAdmin()) return;

  bindLogout();
  highlightActiveNav();

  document.getElementById("categoryForm")?.addEventListener("submit", handleCategorySubmit);

  document.getElementById("categoryName")?.addEventListener("input", (event) => {
    if (slugTouched) return;
    const slugInput = document.getElementById("categorySlug");
    if (slugInput) slugInput.value = slugify(event.target.value);
  });

  document.getElementById("categorySlug")?.addEventListener("input", () => {
    slugTouched = true;
  });

  try {
    if (getFormMode() === "edit") {
      await loadCategoryForEdit();
    }
  } catch (error) {
    showPageError(error);
  }
});
