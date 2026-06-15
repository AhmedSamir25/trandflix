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
    const res = await fetch("/admin/subscription-plan", {
      headers: { Authorization: `Bearer ${token}`, Accept: "application/json" },
    });
    const data = await res.json().catch(() => ({}));
    if (!res.ok) {
      showPageError(data?.msg || "Failed to load plan");
      return;
    }

    const plan = data?.plan;
    if (plan) {
      nameInput.value = plan.name || "";
      priceInput.value = plan.price || "";
      currencyInput.value = plan.currency || "";
      billingDaysInput.value = plan.billing_period_days || 30;
      isActiveInput.checked = plan.is_active === true;
    }
  } catch (err) {
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
