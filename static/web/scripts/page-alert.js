// Allow HTMX error responses to apply out-of-band swaps for #page-alert while
// leaving the original request target untouched (HX-Reswap: none).
document.body.addEventListener("htmx:beforeSwap", (event) => {
  const status = event.detail.xhr && event.detail.xhr.status;
  if (!status || status < 400) {
    return;
  }
  event.detail.shouldSwap = true;
  event.detail.isError = true;
});
