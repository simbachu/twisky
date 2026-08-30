document.addEventListener("click", (event) => {
  const trigger = event.target.closest("[data-compose-open]");
  if (!(trigger instanceof HTMLAnchorElement)) {
    return;
  }
  const dialog = document.getElementById("compose-dialog");
  if (!(dialog instanceof HTMLDialogElement)) {
    return;
  }
  event.preventDefault();
  dialog.showModal();
  const textarea = dialog.querySelector('textarea[name="text"]');
  if (textarea instanceof HTMLTextAreaElement) {
    textarea.focus();
  }
});
