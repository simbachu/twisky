(function () {
  function init() {
    const feedList = document.getElementById("feed-list");
    const toTop = document.getElementById("feed-to-top");
    if (!feedList || !toTop) return;

    const countEl = toTop.querySelector(".feed-to-top-count");
    const header = document.querySelector("main > header");

    function syncCount() {
      const banner = document.querySelector("#new-posts-slot button[data-new-post-count]");
      const raw = banner ? banner.getAttribute("data-new-post-count") : "";
      const count = raw ? Number.parseInt(raw, 10) : 0;
      if (!countEl) return;

      if (!Number.isFinite(count) || count <= 0) {
        countEl.textContent = "";
        countEl.hidden = true;
        toTop.setAttribute("aria-label", "Back to top");
        return;
      }

      countEl.textContent = raw;
      countEl.hidden = false;
      const noun = count === 1 ? "post" : "posts";
      toTop.setAttribute("aria-label", `Back to top, ${raw} new ${noun}`);
    }

    function setVisible(visible) {
      toTop.hidden = !visible;
    }

    if (header && "IntersectionObserver" in window) {
      const observer = new IntersectionObserver(
        (entries) => {
          const entry = entries[0];
          if (!entry) return;
          setVisible(!entry.isIntersecting);
        },
        { threshold: 0 }
      );
      observer.observe(header);
    }

    toTop.addEventListener("click", () => {
      window.scrollTo({ top: 0, behavior: "smooth" });
      const banner = document.querySelector("#new-posts-slot button");
      if (banner instanceof HTMLButtonElement) {
        banner.click();
      }
    });

    document.body.addEventListener("htmx:afterSwap", (event) => {
      const target = event.detail && event.detail.target;
      if (!target) return;
      if (target.id === "new-posts-slot" || target.id === "feed-list") {
        syncCount();
      }
    });

    syncCount();
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }
})();
