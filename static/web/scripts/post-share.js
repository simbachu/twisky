function absoluteCopyURL(rawURL) {
  try {
    return new URL(rawURL, window.location.origin).href;
  } catch {
    return rawURL;
  }
}

function shareGroupFromTarget(target) {
  if (!(target instanceof Element)) return null;
  return target.closest(".post-share-group");
}

function setShareGroupOpen(group, open) {
  const trigger = group.querySelector(".post-share-open");
  group.classList.toggle("post-share-group--open", open);
  if (trigger instanceof HTMLButtonElement) {
    trigger.setAttribute("aria-expanded", open ? "true" : "false");
  }
}

function closeShareGroups(except) {
  for (const group of document.querySelectorAll(".post-share-group")) {
    if (group !== except) {
      setShareGroupOpen(group, false);
    }
  }
}

async function copyShareURL(button) {
  const rawURL = button.dataset.copyUrl;
  if (!rawURL) return;

  const url = absoluteCopyURL(rawURL);
  await navigator.clipboard.writeText(url);

  const label = button.getAttribute("aria-label") || "";
  button.setAttribute("aria-label", "Copied");
  if (button.dataset.copyFeedback === "icon") {
    button.classList.add("ui-icon-engaged");
  }
  window.setTimeout(() => {
    button.setAttribute("aria-label", label);
    button.classList.remove("ui-icon-engaged");
  }, 1500);

  const group = shareGroupFromTarget(button);
  if (group) {
    setShareGroupOpen(group, false);
  }
}

function onShareOpenClick(event) {
  const button = event.target.closest(".post-share-open");
  if (!(button instanceof HTMLButtonElement)) return;

  const group = shareGroupFromTarget(button);
  if (!group) return;

  event.preventDefault();
  event.stopPropagation();

  const open = !group.classList.contains("post-share-group--open");
  closeShareGroups(open ? group : null);
  setShareGroupOpen(group, open);
}

function onShareCopyClick(event) {
  const button = event.target.closest(".post-share-group [data-copy-url]");
  if (!(button instanceof HTMLButtonElement)) return;

  event.preventDefault();
  event.stopPropagation();
  copyShareURL(button).catch(() => {});
}

function onDocumentClick(event) {
  const target = event.target;
  if (!(target instanceof Element)) return;
  if (target.closest(".post-share-group")) return;
  closeShareGroups(null);
}

document.addEventListener("click", onShareOpenClick);
document.addEventListener("click", onShareCopyClick);
document.addEventListener("click", onDocumentClick);
