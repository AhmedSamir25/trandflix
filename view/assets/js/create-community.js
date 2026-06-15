function showFormMsg(message, ok) {
  const el = document.getElementById("formMsg");
  if (!el) return;
  el.hidden = false;
  el.textContent = message;
  el.className = "comm-form-msg " + (ok ? "ok" : "err");
}

function showPendingReview(community) {
  const el = document.getElementById("formMsg");
  if (!el) return;
  const slug = community?.slug || "";
  const viewLink = slug
    ? `/pages/community.html?slug=${encodeURIComponent(slug)}`
    : "/pages/communities.html";
  el.hidden = false;
  el.className = "comm-form-msg comm-review-notice";
  el.innerHTML = `
    <strong>⏳ ${commEscapeHtml(commT("communities.ccCreated"))}</strong>
    <span>${commEscapeHtml(commT("communities.ccPendingReview"))}</span>
    <a class="comm-btn comm-btn-primary" href="${viewLink}">${commEscapeHtml(commT("communities.ccViewCommunity"))}</a>
  `;
  el.scrollIntoView({ behavior: "smooth", block: "center" });
}

window.addEventListener("DOMContentLoaded", () => {
  if (commRedirectIfUnauth()) return;

  const form = document.getElementById("createForm");
  if (!form) return;

  form.addEventListener("submit", async (e) => {
    e.preventDefault();
    const btn = document.getElementById("submitBtn");
    btn.disabled = true;
    showFormMsg(commT("communities.ccCreating"), true);

    const payload = {
      name: document.getElementById("name").value.trim(),
      category_type: document.getElementById("category_type").value,
      avatar_image: document.getElementById("avatar_image").value.trim(),
      cover_image: document.getElementById("cover_image").value.trim(),
      description: document.getElementById("description").value.trim(),
      rules: document.getElementById("rules").value.trim(),
      is_private: document.getElementById("is_private").checked,
    };

    try {
      const res = await commFetch("/api/communities", {
        method: "POST",
        body: JSON.stringify(payload),
      });
      const community = res?.data?.community;
      showPendingReview(community);
    } catch (err) {
      showFormMsg(err.message || commT("communities.ccFailed"), false);
      btn.disabled = false;
    }
  });
});
