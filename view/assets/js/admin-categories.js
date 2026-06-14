let categories = [];

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
             <strong>${escapeHtml(categoryName(c))}</strong>
             <code>${escapeHtml(c.slug)}</code>
           </span>
          <div class="row-actions">
            <a class="text-btn compact-action" href="/pages/admin/edit-category.html?id=${c.id}">${escapeHtml(t("admin.editCategory"))}</a>
            <button class="text-btn danger-btn compact-action" type="button" data-delete-category="${c.id}">${escapeHtml(t("admin.deleteCategory"))}</button>
          </div>
        </article>`,
    )
    .join("");
}

async function loadCategories() {
  clearNotice("pageError");
  const data = await fetchJson("/categories");
  categories = Array.isArray(data?.categories) ? data.categories : [];
  renderCategoryList();
}

async function deleteCategory(categoryId) {
  const category = categories.find((entry) => Number(entry.id) === Number(categoryId));
  const name = categoryName(category) || t("admin.thisCategory");
  if (!window.confirm(t("admin.confirmDeleteCategory").replace("{name}", name))) return;

  clearNotice("pageError");

  const response = await fetchJson(`/categories/${categoryId}`, {
    method: "DELETE",
    headers: authHeaders(),
  });

  setNotice("pageError", response?.msg || t("admin.categoryDeleted"), "success");
  await loadCategories();
}

window.addEventListener("DOMContentLoaded", async () => {
  if (!requireAdmin()) return;

  bindLogout();
  highlightActiveNav();

  document.getElementById("categoryList")?.addEventListener("click", (event) => {
    const deleteBtn = event.target.closest("[data-delete-category]");
    if (deleteBtn) {
      deleteCategory(deleteBtn.dataset.deleteCategory).catch(showPageError);
    }
  });

  window.addEventListener("trendflix:languagechange", renderCategoryList);

  try {
    await loadCategories();
  } catch (error) {
    showPageError(error);
  }
});
