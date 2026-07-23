const DEFAULT_INTERVAL_MS = 10000;
const DEFAULT_MAX_INTERVAL_MS = 5 * 60 * 1000;
const DEFAULT_BURST_INTERVAL_MS = 5000;
const DEFAULT_REPLIES_COOLDOWN_MS = 20000;

// State is keyed by post id so it survives OOB replacement of the poller
// element (e.g. when CountsRefreshFragment refreshes cooldown attrs).
const pollerState = new Map();

function postIdFromPollerId(el) {
  return el.id.replace(/^counts-poller-/, "");
}

function getState(postId) {
  return pollerState.get(postId);
}

function clearScheduled(postId) {
  const state = getState(postId);
  if (state && state.timerId) {
    clearTimeout(state.timerId);
    state.timerId = null;
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

function burstInterval(el) {
  return parseInt(el.dataset.burstIntervalMs, 10) || DEFAULT_BURST_INTERVAL_MS;
}

function repliesCooldown(el) {
  return parseInt(el.dataset.repliesCooldownMs, 10) || DEFAULT_REPLIES_COOLDOWN_MS;
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

function readCounts(el) {
  const postId = postIdFromPollerId(el);
  const counts = {};
  ["like", "reply", "repost"].forEach((metric) => {
    const span = document.getElementById(`${metric}-count-${postId}`);
    if (!span || span.title === "") return;
    const n = parseInt(span.title, 10);
    if (!Number.isNaN(n)) counts[metric] = n;
  });
  return counts;
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
  const postId = postIdFromPollerId(el);
  clearScheduled(postId);
  if (!isLive(el)) return;
  const state = getState(postId);
  if (!state) return;
  state.timerId = setTimeout(() => tick(el), delayMs);
}

function knownReplyIDs(postId) {
  const container = document.getElementById(`post-replies-${postId}`);
  if (!container) return [];
  return Array.from(container.querySelectorAll('[id^="post-"]'))
    .map((node) => node.id.replace(/^post-/, ""))
    .filter(Boolean);
}

function repliesRequestURL(el) {
  const href = el.dataset.repliesHref;
  if (!href) return null;
  const postId = postIdFromPollerId(el);
  const known = knownReplyIDs(postId);
  if (known.length === 0) return href;
  const sep = href.includes("?") ? "&" : "?";
  return `${href}${sep}known=${encodeURIComponent(known.join(","))}`;
}

function maybeFetchReplies(el, state, prev, next) {
  // Reply loading is event-driven off an exact reply-count increase (title
  // attribute), not fuzzy display. When a counts OOB omits the reply span
  // because fuzzy display is unchanged, we intentionally skip (v1).
  if (prev.reply === undefined || next.reply === undefined) return;
  if (next.reply <= prev.reply) return;
  if (!isLive(el) || document.visibilityState !== "visible") return;
  if (Date.now() - state.lastRepliesFetchAt < repliesCooldown(el)) return;

  const url = repliesRequestURL(el);
  if (!url) return;

  htmx.ajax("GET", url, { source: el, swap: "none" }).then(
    () => {
      state.lastRepliesFetchAt = Date.now();
    },
    () => {
      // Leave lastRepliesFetchAt unchanged so the next reply-count bump can retry.
    }
  );
}

function nextIntervalAfterCounts(el, failed, prev, next) {
  if (failed) {
    const state = getState(postIdFromPollerId(el));
    const current = (state && state.currentIntervalMs) || baseInterval(el);
    return Math.min(current * 2, maxInterval(el));
  }
  const replyUp =
    prev.reply !== undefined && next.reply !== undefined && next.reply > prev.reply;
  const repostUp =
    prev.repost !== undefined && next.repost !== undefined && next.repost > prev.repost;
  if (replyUp || repostUp) {
    return Math.min(baseInterval(el), burstInterval(el));
  }
  return baseInterval(el);
}

function tick(el) {
  if (!isLive(el)) return;
  if (document.visibilityState !== "visible") {
    // Left un-scheduled; the visibilitychange handler resumes polling.
    return;
  }

  const postId = postIdFromPollerId(el);
  const state = getState(postId);
  if (!state) return;

  const url = requestURL(el);
  if (!url) return;

  const prev = state.lastCounts;
  let failed = false;
  const onResponseError = () => {
    failed = true;
  };
  el.addEventListener("htmx:responseError", onResponseError, { once: true });
  el.addEventListener("htmx:sendError", onResponseError, { once: true });

  const settle = () => {
    el.removeEventListener("htmx:responseError", onResponseError);
    el.removeEventListener("htmx:sendError", onResponseError);

    // Prefer the live poller node in case an OOB swap replaced el mid-request.
    const liveEl = document.getElementById(el.id) || el;
    const next = readCounts(liveEl);
    if (!failed) {
      maybeFetchReplies(liveEl, state, prev, next);
      state.lastCounts = next;
    }

    const delay = nextIntervalAfterCounts(liveEl, failed, prev, next);
    state.currentIntervalMs = delay;
    liveEl.dataset.currentIntervalMs = String(delay);
    scheduleTick(liveEl, delay);
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
  const postId = postIdFromPollerId(el);
  if (!isLive(el)) {
    clearScheduled(postId);
    pollerState.delete(postId);
    return;
  }

  let state = getState(postId);
  if (state) {
    // Poller OOB-replaced while live: keep burst/cooldown state, rebind timer
    // to the new element. Reschedule at the current interval (may fire early
    // vs remaining delay, but preserves burst mode).
    clearScheduled(postId);
    state.el = el;
    el.dataset.currentIntervalMs = String(state.currentIntervalMs);
    scheduleTick(el, state.currentIntervalMs);
    return;
  }

  state = {
    el,
    lastRepliesFetchAt: 0,
    lastCounts: readCounts(el),
    currentIntervalMs: baseInterval(el),
    timerId: null,
  };
  pollerState.set(postId, state);
  el.dataset.currentIntervalMs = String(state.currentIntervalMs);
  scheduleTick(el, state.currentIntervalMs);
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
    clearScheduled(postIdFromPollerId(el));
    tick(el);
  });
});
