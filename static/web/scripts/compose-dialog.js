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
  prepareComposeDialog(dialog, trigger);
  dialog.showModal();
  const textarea = dialog.querySelector('textarea[name="text"]');
  if (textarea instanceof HTMLTextAreaElement) {
    textarea.focus();
  }
});

document.addEventListener("close", (event) => {
  const dialog = event.target;
  if (!(dialog instanceof HTMLDialogElement) || dialog.id !== "compose-dialog") {
    return;
  }
  resetComposeDialog(dialog);
}, true);

function prepareComposeDialog(dialog, trigger) {
  const title = dialog.querySelector("#compose-dialog-title");
  const slot = dialog.querySelector("#compose-parent-slot");
  const parentInput = dialog.querySelector('input[name="parent"]');
  const parentParam = parentFromHref(trigger.href);

  clearParentSlot(slot);
  if (parentInput instanceof HTMLInputElement) {
    parentInput.value = parentParam;
  }
  if (title) {
    title.textContent = parentParam ? "Reply" : "New post";
  }
  if (!parentParam || !(slot instanceof HTMLElement)) {
    return;
  }
  const template = trigger.parentElement?.querySelector("template.compose-parent-template");
  if (!(template instanceof HTMLTemplateElement)) {
    return;
  }
  slot.appendChild(template.content.cloneNode(true));
}

function resetComposeDialog(dialog) {
  const title = dialog.querySelector("#compose-dialog-title");
  const slot = dialog.querySelector("#compose-parent-slot");
  const parentInput = dialog.querySelector('input[name="parent"]');
  const form = dialog.querySelector("form.compose-field");
  clearParentSlot(slot);
  if (parentInput instanceof HTMLInputElement) {
    parentInput.value = "";
  }
  if (title) {
    title.textContent = "New post";
  }
  if (form instanceof HTMLFormElement) {
    form.reset();
  }
}

function clearParentSlot(slot) {
  if (!(slot instanceof HTMLElement)) {
    return;
  }
  while (slot.firstChild) {
    slot.removeChild(slot.firstChild);
  }
}

function parentFromHref(href) {
  try {
    return new URL(href, window.location.origin).searchParams.get("parent") || "";
  } catch {
    return "";
  }
}
