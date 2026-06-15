function showFormMsg(message, ok) {
  const el = document.getElementById("formMsg");
  if (!el) return;
  el.hidden = false;
  el.textContent = message;
  el.className = "comm-form-msg " + (ok ? "ok" : "err");
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
      if (community?.slug) {
        window.location.href = `/pages/community.html?slug=${encodeURIComponent(community.slug)}`;
      } else {
        showFormMsg(commT("communities.ccCreated"), true);
      }
    } catch (err) {
      showFormMsg(err.message || commT("communities.ccFailed"), false);
      btn.disabled = false;
    }
  });
});
