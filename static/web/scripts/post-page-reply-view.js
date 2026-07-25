const REPLY_VIEW_COOKIE = "twisky-reply-view";
const REPLY_VIEW_LINEAR = "linear";
const REPLY_VIEW_THREADED = "threaded";
const REPLY_VIEW_LINEAR_CLASS = "reply-view-linear";
const REPLY_VIEW_COOKIE_MAX_AGE = 31536000;

function readReplyViewCookie() {
  const match = document.cookie.match(
    new RegExp(`(?:^|; )${REPLY_VIEW_COOKIE}=([^;]*)`)
  );
  if (!match) return REPLY_VIEW_THREADED;
  const value = decodeURIComponent(match[1]);
  return value === REPLY_VIEW_LINEAR ? REPLY_VIEW_LINEAR : REPLY_VIEW_THREADED;
}

function writeReplyViewCookie(view) {
  document.cookie = `${REPLY_VIEW_COOKIE}=${encodeURIComponent(view)}; Path=/; Max-Age=${REPLY_VIEW_COOKIE_MAX_AGE}; SameSite=Lax`;
}

function rootRepliesList() {
  return document.querySelector(".post-page > ul.post-replies[id]");
}

function syncDocumentReplyView(view) {
  if (view === REPLY_VIEW_LINEAR) {
    document.documentElement.dataset.replyView = REPLY_VIEW_LINEAR;
  } else {
    delete document.documentElement.dataset.replyView;
  }
}

function applyReplyView(view) {
  syncDocumentReplyView(view);
  const root = rootRepliesList();
  if (root) {
    root.classList.toggle(REPLY_VIEW_LINEAR_CLASS, view === REPLY_VIEW_LINEAR);
  }
  const threaded = document.getElementById("reply-view-threaded");
  const linear = document.getElementById("reply-view-linear");
  if (threaded) threaded.checked = view === REPLY_VIEW_THREADED;
  if (linear) linear.checked = view === REPLY_VIEW_LINEAR;
}

function initReplyView() {
  if (!document.getElementById("post-page-header")) return;
  applyReplyView(readReplyViewCookie());
}

function onReplyViewChange(event) {
  const input = event.target;
  if (!(input instanceof HTMLInputElement)) return;
  if (input.name !== "reply-view" || !input.checked) return;
  const view =
    input.value === REPLY_VIEW_LINEAR ? REPLY_VIEW_LINEAR : REPLY_VIEW_THREADED;
  writeReplyViewCookie(view);
  applyReplyView(view);
}

function onRepliesOOBAfterSwap(event) {
  const target = event.detail.target;
  if (!(target instanceof HTMLElement)) return;
  if (!target.matches('ul.post-replies[id^="post-replies-"]')) return;
  if (!target.closest(".post-page")) return;
  target.classList.toggle(
    REPLY_VIEW_LINEAR_CLASS,
    readReplyViewCookie() === REPLY_VIEW_LINEAR
  );
}

document.addEventListener("DOMContentLoaded", initReplyView);
document.addEventListener("change", onReplyViewChange);
document.body.addEventListener("htmx:oobAfterSwap", onRepliesOOBAfterSwap);
