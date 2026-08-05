function sameOriginReferrer() {
  const referrer = document.referrer;
  if (!referrer) return false;
  try {
    return new URL(referrer).origin === window.location.origin;
  } catch {
    return false;
  }
}

document.addEventListener("click", (event) => {
  const link = event.target instanceof Element ? event.target.closest("[data-nav-back]") : null;
  if (!(link instanceof HTMLAnchorElement)) return;
  if (event.defaultPrevented || event.button !== 0) return;
  if (event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return;
  if (!sameOriginReferrer()) return;

  event.preventDefault();
  history.back();
});
