(function () {
  const defaultIcon = "/static/icons/favicon.png";
  const notifyIcon = "/static/icons/favicon-notify.png";

  function hasPendingNewPosts() {
    return document.querySelector("#new-posts-slot .new-posts-button") !== null;
  }

  function updateFavicon() {
    const link = document.getElementById("page-favicon");
    if (!link) return;
    link.href = hasPendingNewPosts() && document.hidden ? notifyIcon : defaultIcon;
  }

  document.addEventListener("visibilitychange", updateFavicon);

  document.addEventListener("DOMContentLoaded", function () {
    document.body.addEventListener("htmx:afterSwap", function (event) {
      const target = event.detail.target;
      if (target.id === "new-posts-slot" || target.id === "feed-list") {
        updateFavicon();
      }
    });
  });
})();
