function communityIdFromUrl() {
  return new URLSearchParams(window.location.search).get("community") || "";
}

function showFormMsg(message, ok) {
  const el = document.getElementById("formMsg");
  if (!el) return;
  el.hidden = false;
  el.textContent = message;
  el.className = "comm-form-msg " + (ok ? "ok" : "err");
}

let currentCommunityId = "";

async function loadCommunityInfo() {
  currentCommunityId = communityIdFromUrl();
  const subtitle = document.getElementById("subtitle");
  const cancelLink = document.getElementById("cancelLink");

  if (!currentCommunityId) {
    subtitle.textContent = commT("communities.cpNoCommunity");
    return;
  }

  try {
    const res = await commFetch(`/api/communities/${encodeURIComponent(currentCommunityId)}/posts?per_page=1`);
    // posts endpoint works without slug; we just confirm the community exists.
    cancelLink.href = `/pages/community.html`;
  } catch (err) {
    subtitle.textContent = err.message || commT("communities.cpFailedLoad");
  }
}

window.addEventListener("DOMContentLoaded", async () => {
  if (commRedirectIfUnauth()) return;
  await loadCommunityInfo();

  const form = document.getElementById("postForm");
  form.addEventListener("submit", async (e) => {
    e.preventDefault();
    if (!currentCommunityId) {
      showFormMsg(commT("communities.cpNoCommunity"), false);
      return;
    }
    const btn = document.getElementById("submitBtn");
    btn.disabled = true;
    showFormMsg(commT("communities.cpPublishing"), true);

    const relatedType = document.getElementById("related_item_type").value.trim();
    const payload = {
      title: document.getElementById("title").value.trim(),
      body: document.getElementById("body").value.trim(),
      post_type: document.getElementById("post_type").value,
      is_spoiler: document.getElementById("is_spoiler").checked,
    };
    if (relatedType) {
      payload.related_item_type = relatedType;
    }

    try {
      const res = await commFetch(`/api/communities/${currentCommunityId}/posts`, {
        method: "POST",
        body: JSON.stringify(payload),
      });
      const post = res?.data?.post;
      if (post?.id) {
        window.location.href = `/pages/post.html?id=${encodeURIComponent(post.id)}`;
      } else {
        showFormMsg(commT("communities.cpPublished"), true);
      }
    } catch (err) {
      showFormMsg(err.message || commT("communities.cpPublishFailed"), false);
      btn.disabled = false;
    }
  });
});
