const DEFAULT_INTERVAL_MS = 10000;
const DEFAULT_MAX_INTERVAL_MS = 5 * 60 * 1000;

const timers = new WeakMap();

function clearScheduled(el) {
  const timerId = timers.get(el);
  if (timerId) {
    clearTimeout(timerId);
    timers.delete(el);
  }
}

function isLive(el) {
  return el.dataset.live === "true";
}

function baseInterval(el) {
  return parseInt(el.dataset.intervalMs, 10) || DEFAULT_INTERVAL_MS;
}

function maxInterval(el) {
  return parseInt(el.dataset.maxIntervalMs, 10) || DEFAULT_MAX_INTERVAL_MS;
}

function currentInterval(el) {
  return parseInt(el.dataset.currentIntervalMs, 10) || baseInterval(el);
}

function postIdFromPollerId(el) {
  return el.id.replace(/^counts-poller-/, "");
}

function currentCountParams(el) {
  const postId = postIdFromPollerId(el);
  const params = new URLSearchParams();
  ["like", "reply", "repost"].forEach((metric) => {
    const span = document.getElementById(`${metric}-count-${postId}`);
    if (span && span.title !== "") {
      params.set(metric, span.title);
    }
  });
  return params;
}

function requestURL(el) {
  const href = el.dataset.href;
  if (!href) return null;
  const params = currentCountParams(el);
  if ([...params].length === 0) return href;
  const sep = href.includes("?") ? "&" : "?";
  return `${href}${sep}${params.toString()}`;
}

function scheduleTick(el, delayMs) {
  clearScheduled(el);
  if (!isLive(el)) return;
  timers.set(
    el,
    setTimeout(() => tick(el), delayMs)
  );
}

function tick(el) {
  if (!isLive(el)) return;
  if (document.visibilityState !== "visible") {
    // Left un-scheduled; the visibilitychange handler resumes polling.
    return;
  }

  const url = requestURL(el);
  if (!url) return;

  let failed = false;
  const onResponseError = () => {
    failed = true;
  };
  el.addEventListener("htmx:responseError", onResponseError, { once: true });
  el.addEventListener("htmx:sendError", onResponseError, { once: true });

  const settle = () => {
    el.removeEventListener("htmx:responseError", onResponseError);
    el.removeEventListener("htmx:sendError", onResponseError);

    const next = failed
      ? Math.min(currentInterval(el) * 2, maxInterval(el))
      : baseInterval(el);
    el.dataset.currentIntervalMs = String(next);
    scheduleTick(el, next);
  };

  // htmx.ajax() resolves after the request settles, even on an HTTP error
  // response (those surface via htmx:responseError instead) - but a rejected
  // promise (e.g. a network-level failure) should still count as a failure
  // and back off rather than silently stop polling.
  htmx.ajax("GET", url, { source: el, swap: "none" }).then(settle, () => {
    failed = true;
    settle();
  });
}

function initPoller(el) {
  clearScheduled(el);
  if (!isLive(el)) return;
  el.dataset.currentIntervalMs = String(baseInterval(el));
  scheduleTick(el, baseInterval(el));
}

function initPollersWithin(root) {
  if (!root || typeof root.querySelectorAll !== "function") return;
  if (root.matches && root.matches("[data-counts-poll]")) {
    initPoller(root);
  }
  root.querySelectorAll("[data-counts-poll]").forEach(initPoller);
}

document.addEventListener("DOMContentLoaded", () => initPollersWithin(document));

document.body.addEventListener("htmx:afterSwap", (event) => {
  initPollersWithin(event.detail.target);
});

document.body.addEventListener("htmx:oobAfterSwap", (event) => {
  initPollersWithin(event.detail.target);
});

document.addEventListener("visibilitychange", () => {
  if (document.visibilityState !== "visible") return;
  document.querySelectorAll('[data-counts-poll][data-live="true"]').forEach((el) => {
    clearScheduled(el);
    tick(el);
  });
});
