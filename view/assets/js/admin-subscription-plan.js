const TOKEN_KEY = "trendflix.token";

function requireAdmin() {
  const token = localStorage.getItem(TOKEN_KEY);
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
  const el = document.getElementById("planStatus");
  if (!el) return;
  el.textContent = msg;
  el.classList.toggle("error-notice", isError);
  el.classList.toggle("success-notice", !isError);
  el.hidden = !msg;
}

function setPlanLoading(isLoading, enableControls = false) {
  const form = document.getElementById("planForm");
  const loading = document.getElementById("planLoading");

  if (form) {
    form.setAttribute("aria-busy", isLoading ? "true" : "false");
    form.querySelectorAll("input, select, textarea, button").forEach((el) => {
      el.disabled = isLoading ? true : !enableControls;
    });
  }

  if (loading) {
    loading.hidden = !isLoading;
  }
}

function populatePlanForm(plan) {
  const nameInput = document.getElementById("planName");
  const priceInput = document.getElementById("planPrice");
  const currencyInput = document.getElementById("planCurrency");
  const billingDaysInput = document.getElementById("planBillingDays");
  const isActiveInput = document.getElementById("planIsActive");

  if (!plan) return;

  if (nameInput) nameInput.value = plan.name ?? "";
  if (priceInput) priceInput.value = plan.price ?? "";
  if (currencyInput) currencyInput.value = plan.currency ?? "";
  if (billingDaysInput) billingDaysInput.value = plan.billing_period_days ?? "";
  if (isActiveInput) isActiveInput.checked = plan.is_active === true;
}

window.addEventListener("DOMContentLoaded", async () => {
  const token = requireAdmin();
  if (!token) return;

  const nameInput = document.getElementById("planName");
  const priceInput = document.getElementById("planPrice");
  const currencyInput = document.getElementById("planCurrency");
  const billingDaysInput = document.getElementById("planBillingDays");
  const isActiveInput = document.getElementById("planIsActive");
  const form = document.getElementById("planForm");
  const saveBtn = document.getElementById("savePlanBtn");

  try {
    setPlanLoading(true);
    const res = await fetch("/admin/subscription-plan", {
      headers: { Authorization: `Bearer ${token}`, Accept: "application/json" },
    });
    const data = await res.json().catch(() => ({}));
    if (!res.ok) {
      setPlanLoading(false, false);
      showPageError(data?.msg || "Failed to load plan");
      return;
    }

    const plan = data?.plan ?? data?.subscription_plan ?? null;
    populatePlanForm(plan);
    setPlanLoading(false, true);
  } catch (err) {
    setPlanLoading(false, false);
    showPageError(err?.message || "Failed to load plan");
    return;
  }

  form.addEventListener("submit", async (e) => {
    e.preventDefault();
    showStatus("");
    saveBtn.disabled = true;

    const body = {
      name: String(nameInput.value || "").trim(),
      price: parseFloat(priceInput.value) || 0,
      currency: String(currencyInput.value || "").trim().toUpperCase(),
      billing_period_days: parseInt(billingDaysInput.value, 10) || 30,
      is_active: isActiveInput.checked,
    };

    try {
      const res = await fetch("/admin/subscription-plan", {
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
        showStatus(data?.msg || "Failed to save plan", true);
        return;
      }
      showStatus("Plan updated successfully", false);
    } catch (err) {
      showStatus(err?.message || "Failed to save plan", true);
    } finally {
      saveBtn.disabled = false;
    }
  });
});
