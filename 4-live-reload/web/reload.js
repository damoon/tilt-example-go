
var liveReload = function () {
    fetch('/live-reload?' + new URLSearchParams({currentBuild: currentBuild}))
    .then((response) => response.json())
    .then((response) => {
        if (response.disabled) {
            return;
        }
        
        delay = 1000;
        if (currentBuild == "") {
            delay = 0;
        }
        
        currentBuild = response.build;
        if (response.reload) {
            window.location.reload();
        }

        window.setTimeout(liveReload, delay);
    })
    .catch((error) => {
      console.error('Error:', error);
      window.setTimeout(liveReload, 1000);
    });
}
window.setTimeout(liveReload, 1);
