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

function scrollToPostPageHeader() {
  const header = document.getElementById("post-page-header");
  if (!header) return;

  let targetScrollY = header.getBoundingClientRect().top + window.scrollY;
  let maxScrollY =
    document.documentElement.scrollHeight - window.innerHeight;

  if (targetScrollY > maxScrollY) {
    const spacer = scrollSpacer();
    if (spacer) {
      spacer.style.height = `${targetScrollY - maxScrollY}px`;
      void document.documentElement.offsetHeight;
      maxScrollY =
        document.documentElement.scrollHeight - window.innerHeight;
    }
  }

  window.scrollTo(0, Math.min(targetScrollY, maxScrollY));
  header.focus({ preventScroll: true });
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
  if (event.detail.target.id !== "post-page-ancestors") return;
  scrollToPostPageHeader();
});
