const AI_TOKEN_KEY = "trendflix.token";

function requireAdmin() {
  const token = localStorage.getItem(AI_TOKEN_KEY);
  if (!token) {
    window.location.replace("/pages/auth/auth.html");
    return null;
  }

  try {
    const payload = token.split(".")[1] || "";
    const normalized = payload.replaceAll("-", "+").replaceAll("_", "/");
    const padded = normalized.padEnd(Math.ceil(normalized.length / 4) * 4, "=");
    const parsed = JSON.parse(window.atob(padded));
    if (String(parsed?.role || "").trim().toLowerCase() !== "admin") {
      window.location.replace("/pages/app.html");
      return null;
    }
  } catch {
    window.location.replace("/pages/auth/auth.html");
    return null;
  }

  return token;
}

function showPageError(msg) {
  const el = document.getElementById("pageError");
  if (!el) return;
  el.textContent = msg;
  el.hidden = !msg;
}

function showStatus(msg, isError = false) {
  const el = document.getElementById("aiStatus");
  if (!el) return;
  el.textContent = msg;
  el.classList.toggle("error-notice", isError);
  el.classList.toggle("success-notice", !isError);
  el.hidden = !msg;
}

function syncProviderGroups() {
  const provider = document.getElementById("aiProvider").value;
  const orGroup = document.getElementById("openRouterGroup");
  const oaiGroup = document.getElementById("openAICompatibleGroup");
  if (!orGroup || !oaiGroup) return;
  orGroup.hidden = provider !== "openrouter";
  oaiGroup.hidden = provider !== "openai_compatible";
}

window.addEventListener("DOMContentLoaded", async () => {
  const token = requireAdmin();
  if (!token) return;

  const providerSelect = document.getElementById("aiProvider");
  const openRouterKeyInput = document.getElementById("openRouterKey");
  const openRouterModelInput = document.getElementById("openRouterModel");
  const openAIKeyInput = document.getElementById("openAICompatibleKey");
  const openAIBaseURLInput = document.getElementById("openAICompatibleBaseURL");
  const openAIModelInput = document.getElementById("openAICompatibleModel");
  const form = document.getElementById("aiForm");
  const saveBtn = document.getElementById("saveAiBtn");

  providerSelect.addEventListener("change", syncProviderGroups);

  try {
    const res = await fetch("/admin/ai-settings", {
      headers: { Authorization: `Bearer ${token}`, Accept: "application/json" },
    });
    const data = await res.json().catch(() => ({}));
    if (!res.ok) {
      showPageError(data?.msg || "Failed to load AI settings");
      syncProviderGroups();
      return;
    }

    const settings = data?.settings || {};
    providerSelect.value = settings.provider || "openai_compatible";
    openRouterKeyInput.value = settings.openrouter_api_key || "";
    openRouterModelInput.value = settings.openrouter_model || "";
    openAIKeyInput.value = settings.openai_compatible_api_key || "";
    openAIBaseURLInput.value = settings.openai_compatible_base_url || "";
    openAIModelInput.value = settings.openai_compatible_model || "";
    syncProviderGroups();
  } catch (err) {
    showPageError(err?.message || "Failed to load AI settings");
    syncProviderGroups();
    return;
  }

  form.addEventListener("submit", async (e) => {
    e.preventDefault();
    showStatus("");
    saveBtn.disabled = true;

    const body = {
      provider: String(providerSelect.value || "").trim(),
      openrouter_api_key: String(openRouterKeyInput.value || "").trim(),
      openrouter_model: String(openRouterModelInput.value || "").trim(),
      openai_compatible_api_key: String(openAIKeyInput.value || "").trim(),
      openai_compatible_base_url: String(openAIBaseURLInput.value || "").trim(),
      openai_compatible_model: String(openAIModelInput.value || "").trim(),
    };

    try {
      const res = await fetch("/admin/ai-settings", {
        method: "PUT",
        headers: {
          Authorization: `Bearer ${token}`,
          "Content-Type": "application/json",
          Accept: "application/json",
        },
        body: JSON.stringify(body),
      });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) {
        showStatus(data?.msg || "Failed to save AI settings", true);
        return;
      }

      const settings = data?.settings || {};
      openRouterKeyInput.value = settings.openrouter_api_key || "";
      openAIKeyInput.value = settings.openai_compatible_api_key || "";
      showStatus("AI settings updated successfully", false);
    } catch (err) {
      showStatus(err?.message || "Failed to save AI settings", true);
    } finally {
      saveBtn.disabled = false;
    }
  });
});
