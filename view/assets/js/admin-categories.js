let categories = [];
let slugTouched = false;
let editingCategoryId = null;

function renderCategoryList() {
  const list = document.getElementById("categoryList");
  if (!list) return;

  if (!categories.length) {
    list.innerHTML = `<p class="notice">${escapeHtml(t("admin.noCategories"))}</p>`;
    return;
  }

  list.innerHTML = categories
    .map(
      (c) =>
        `<article class="category-row" data-category-id="${c.id}">
          <span>
            <strong>${escapeHtml(c.name)}</strong>
            <code>${escapeHtml(c.slug)}</code>
          </span>
          <div class="row-actions">
            <button class="text-btn compact-action" type="button" data-edit-category="${c.id}">${escapeHtml(t("admin.editCategory"))}</button>
            <button class="text-btn danger-btn compact-action" type="button" data-delete-category="${c.id}">${escapeHtml(t("admin.deleteCategory"))}</button>
          </div>
        </article>`,
    )
    .join("");
}

function resetCategoryForm() {
  document.getElementById("categoryForm")?.reset();
  editingCategoryId = null;
  slugTouched = false;
  setButtonLoading("createCategoryBtn", "admin.creatingCategory", "admin.createCategory", false);
}

function startCategoryEdit(categoryId) {
  const category = categories.find((entry) => Number(entry.id) === Number(categoryId));
  if (!category) return;

  editingCategoryId = category.id;
  slugTouched = true;
  const nameInput = document.getElementById("categoryName");
  const slugInput = document.getElementById("categorySlug");
  const submitBtn = document.getElementById("createCategoryBtn");

  if (nameInput) nameInput.value = category.name || "";
  if (slugInput) slugInput.value = category.slug || "";
  if (submitBtn) submitBtn.textContent = t("admin.saveCategory");
}

async function loadCategories() {
  clearNotice("pageError");
  const data = await fetchJson("/categories");
  categories = Array.isArray(data?.categories) ? data.categories : [];
  renderCategoryList();
}

async function handleCategorySubmit(event) {
  event.preventDefault();
  clearNotice("pageError");
  clearNotice("categoryStatus");

  const nameInput = document.getElementById("categoryName");
  const slugInput = document.getElementById("categorySlug");
  const name = String(nameInput?.value || "").trim();
  const slug = slugify(slugInput?.value || name);

  if (!name) {
    setNotice("categoryStatus", t("admin.categoryNameRequired"), "error");
    return;
  }

  if (slugInput) slugInput.value = slug;

  const isEditing = Boolean(editingCategoryId);
  setButtonLoading(
    "createCategoryBtn",
    isEditing ? "admin.savingCategory" : "admin.creatingCategory",
    isEditing ? "admin.saveCategory" : "admin.createCategory",
    true,
  );

  try {
    const response = await fetchJson(isEditing ? `/categories/${editingCategoryId}` : "/categories", {
      method: isEditing ? "PUT" : "POST",
      headers: authHeaders({ "Content-Type": "application/json" }),
      body: JSON.stringify({ name, slug }),
    });

    resetCategoryForm();
    setNotice("categoryStatus", response?.msg || t(isEditing ? "admin.categoryUpdated" : "admin.categoryCreated"), "success");
    await loadCategories();
  } catch (error) {
    setNotice("categoryStatus", error.message, "error");
  } finally {
    const stillEditing = Boolean(editingCategoryId);
    setButtonLoading(
      "createCategoryBtn",
      stillEditing ? "admin.savingCategory" : "admin.creatingCategory",
      stillEditing ? "admin.saveCategory" : "admin.createCategory",
      false,
    );
  }
}

async function deleteCategory(categoryId) {
  const category = categories.find((entry) => Number(entry.id) === Number(categoryId));
  const name = category?.name || t("admin.thisCategory");
  if (!window.confirm(t("admin.confirmDeleteCategory").replace("{name}", name))) return;

  clearNotice("pageError");
  clearNotice("categoryStatus");

  const response = await fetchJson(`/categories/${categoryId}`, {
    method: "DELETE",
    headers: authHeaders(),
  });

  if (Number(editingCategoryId) === Number(categoryId)) resetCategoryForm();
  setNotice("categoryStatus", response?.msg || t("admin.categoryDeleted"), "success");
  await loadCategories();
}

window.addEventListener("DOMContentLoaded", async () => {
  if (!requireAdmin()) return;

  bindLogout();
  highlightActiveNav();

  document.getElementById("categoryForm")?.addEventListener("submit", handleCategorySubmit);

  document.getElementById("refreshCategoriesBtn")?.addEventListener("click", () =>
    loadCategories().catch(showPageError),
  );

  document.getElementById("categoryList")?.addEventListener("click", (event) => {
    const editBtn = event.target.closest("[data-edit-category]");
    if (editBtn) {
      startCategoryEdit(editBtn.dataset.editCategory);
      return;
    }

    const deleteBtn = event.target.closest("[data-delete-category]");
    if (deleteBtn) {
      deleteCategory(deleteBtn.dataset.deleteCategory).catch(showPageError);
    }
  });

  document.getElementById("categoryName")?.addEventListener("input", (event) => {
    if (slugTouched) return;
    const slugInput = document.getElementById("categorySlug");
    if (slugInput) slugInput.value = slugify(event.target.value);
  });

  document.getElementById("categorySlug")?.addEventListener("input", () => {
    slugTouched = true;
  });

  window.addEventListener("trendflix:languagechange", renderCategoryList);

  try {
    await loadCategories();
  } catch (error) {
    showPageError(error);
  }
});
