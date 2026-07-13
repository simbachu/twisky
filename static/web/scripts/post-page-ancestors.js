let focusAnchorTop = null;

function scrollSpacer() {
  let spacer = document.getElementById("post-page-scroll-spacer");
  const main = document.querySelector("main");
  if (!main) return null;
  if (!spacer) {
    spacer = document.createElement("div");
    spacer.id = "post-page-scroll-spacer";
    spacer.setAttribute("aria-hidden", "true");
    main.appendChild(spacer);
  }
  return spacer;
}

function initPostPageAncestors() {
  const slot = document.getElementById("post-page-ancestors");
  if (!slot) return;

  if ("scrollRestoration" in history) {
    history.scrollRestoration = "manual";
  }

  const spacer = scrollSpacer();
  if (spacer) {
    spacer.style.height = "0px";
  }
  window.scrollTo(0, 0);

  const focus = document.querySelector("article.post.post-page");
  focusAnchorTop = focus ? focus.getBoundingClientRect().top : 0;

  htmx.trigger(slot, "twiskyAncestors");
}

document.addEventListener("DOMContentLoaded", initPostPageAncestors);

window.addEventListener("pageshow", (event) => {
  if (!event.persisted || !document.getElementById("post-page-ancestors")) {
    return;
  }
  const slot = document.getElementById("post-page-ancestors");
  slot.innerHTML = "";
  initPostPageAncestors();
});

document.body.addEventListener("htmx:afterSwap", (event) => {
  if (event.detail.target.id !== "post-page-ancestors" || focusAnchorTop === null) return;
  const focus = document.querySelector("article.post.post-page");
  if (!focus) return;

  const delta = focus.getBoundingClientRect().top - focusAnchorTop;
  if (delta === 0) {
    focusAnchorTop = null;
    return;
  }

  let neededScrollY = window.scrollY + delta;
  let maxScrollY =
    document.documentElement.scrollHeight - window.innerHeight;
  if (neededScrollY > maxScrollY) {
    const spacer = scrollSpacer();
    if (spacer) {
      spacer.style.height = `${neededScrollY - maxScrollY}px`;
      void document.documentElement.offsetHeight;
      maxScrollY =
        document.documentElement.scrollHeight - window.innerHeight;
    }
  }

  window.scrollTo(0, Math.min(neededScrollY, maxScrollY));
  focusAnchorTop = null;
});
