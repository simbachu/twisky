(function () {
  var players = [];

  function ignorePlayError(err) {
    if (!err) {
      return;
    }
    if (err.name === "AbortError" || err.name === "NotAllowedError") {
      return;
    }
  }

  function safePlay(video) {
    var result = video.play();
    if (result && typeof result.then === "function") {
      result.catch(ignorePlayError);
    }
  }

  function initPlayer(video) {
    var playlist = video.getAttribute("data-playlist");
    if (!playlist) {
      return;
    }

    var attached = false;
    var shouldPlay = false;

    function hideOverlay() {
      var figure = video.closest(".post-video");
      if (figure) {
        figure.classList.add("post-video--playing");
      }
    }

    function playIfRequested() {
      if (!shouldPlay) {
        return;
      }
      hideOverlay();
      video.setAttribute("controls", "");
      safePlay(video);
    }

    function attach() {
      if (attached) {
        playIfRequested();
        return;
      }
      attached = true;

      if (typeof Hls !== "undefined" && Hls.isSupported()) {
        var hls = new Hls();
        hls.loadSource(playlist);
        hls.attachMedia(video);
        players.push(hls);
        hls.on(Hls.Events.MANIFEST_PARSED, playIfRequested);
        return;
      }

      if (video.canPlayType("application/vnd.apple.mpegurl")) {
        video.addEventListener("canplay", playIfRequested, { once: true });
        video.src = playlist;
        return;
      }
    }

    function requestPlay() {
      shouldPlay = true;
      hideOverlay();
      attach();
    }

    video.addEventListener("click", requestPlay);

    video.addEventListener("play", function () {
      if (!attached) {
        requestPlay();
      }
    });
  }

  function init() {
    var videos = document.querySelectorAll(
      '.post-video-player[data-presentation="default"]'
    );
    for (var i = 0; i < videos.length; i++) {
      initPlayer(videos[i]);
    }
  }

  window.addEventListener("pagehide", function () {
    for (var i = 0; i < players.length; i++) {
      players[i].destroy();
    }
    players = [];
  });

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }
})();
