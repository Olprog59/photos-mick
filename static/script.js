window.initializeMap = function () {
  var mapContainer = document.getElementById("map-container");

  // Recréez complètement l'élément de carte
  var oldMapElement = document.getElementById("map");
  if (oldMapElement) {
    var lat = oldMapElement.getAttribute("data-latitude");
    var lng = oldMapElement.getAttribute("data-longitude");
    var newMapElement = document.createElement("div");
    newMapElement.id = "map";
    newMapElement.setAttribute("data-latitude", lat);
    newMapElement.setAttribute("data-longitude", lng);
    newMapElement.style.height = "400px";
    newMapElement.style.width = "100%";
    mapContainer.innerHTML = "";
    mapContainer.appendChild(newMapElement);
  }

  var mapElement = document.getElementById("map");
  if (!mapElement) {
    console.log("Map element not found");
    return;
  }

  var lat = parseFloat(mapElement.getAttribute("data-latitude"));
  var lng = parseFloat(mapElement.getAttribute("data-longitude"));

  // if (map) {
  //   map.remove();
  // }

  const map = L.map("map").setView([lat, lng], 13);
  var marker = L.marker([lat, lng], { draggable: true }).addTo(map);

  marker.on("dragend", function (e) {
    const { lat, lng } = e.target.getLatLng();
    document.getElementById("latitude").value = lat.toFixed(6);
    document.getElementById("longitude").value = lng.toFixed(6);
  });

  map.on("click", function (e) {
    const { lat, lng } = e.latlng;
    marker.setLatLng([lat, lng]);
    document.getElementById("latitude").value = lat.toFixed(6);
    document.getElementById("longitude").value = lng.toFixed(6);
  });

  L.tileLayer("https://tile.openstreetmap.org/{z}/{x}/{y}.png", {
    maxZoom: 19,
    attribution:
      '&copy; <a href="http://www.openstreetmap.org/copyright">OpenStreetMap</a>',
  }).addTo(map);

  if (L.Control.geocoder) {
    L.Control.geocoder().addTo(map);
  }

  setTimeout(() => {
    if (map) {
      map.invalidateSize();
    }
  }, 100);
};

// footer error
var idTimeout = 0;
const footer = document.getElementById("error_popover");
footer.querySelector("#popover__close").addEventListener("click", () => {
  footer.classList.remove("active");
  clearTimeout(idTimeout);
});

function callback(mutationsList) {
  mutationsList.forEach((mut) => {
    if (mut.attributeName === "class" && mut.target.className === "active") {
      idTimeout = setTimeout(() => {
        mut.target.classList.remove("active");
      }, 6000);
    }
  });
}

const mutationObserver = new MutationObserver(callback);

mutationObserver.observe(footer, { attributes: true });

document.addEventListener("htmx:afterRequest", (evt) => {
  const popover = footer.querySelector("div#popover");

  const failed = evt.detail.failed;

  if (failed) {
    popover.className = "";
    footer.classList.add("active");
    popover.querySelector("p").innerText = evt.detail.xhr.responseText;
    popover.classList.add("error");
  } else {
    const data = evt.detail.xhr.responseText;
    const header = evt.detail.xhr.getResponseHeader("Message");

    if (data && header) {
      popover.className = "";
      footer.classList.add("active");
      popover.querySelector("p").innerText = header;
      popover.classList.add("success");
    }
  }
});

document.addEventListener("htmx:configRequest", function (evt) {
  evt.detail.headers["HX-Trigger"] = "mediaUpdated";
});

document.addEventListener("mediaUpdated", function () {
  htmx.ajax("GET", "/api/medias", "#media-list");
});

document.addEventListener("DOMContentLoaded", function () {
  let currentPage = 1;
  const mediaList = document.getElementById("media-list");

  // Fonction pour mettre à jour l'URL de chargement
  function updateLoadUrl() {
    currentPage++;
    mediaList.setAttribute("hx-get", `/api/medias?page=${currentPage}`);
  }

  // Observateur d'intersection pour le chargement infini
  const observer = new IntersectionObserver(
    (entries) => {
      entries.forEach((entry) => {
        if (entry.isIntersecting) {
          const loadMore = entry.target.querySelector(
            '[hx-trigger="intersect"]',
          );
          if (loadMore) {
            htmx.trigger(loadMore, "intersect");
            updateLoadUrl();
          }
        }
      });
    },
    { rootMargin: "0px 0px 200px 0px" },
  );

  // Observer le conteneur principal
  observer.observe(mediaList);

  // Gérer l'opacité des nouveaux éléments
  mediaList.addEventListener("htmx:afterOnLoad", function () {
    if (document.getElementById("only-change").classList.contains("active")) {
      document
        .querySelectorAll(".media-item:not(.opacity-modified)")
        .forEach((item) => {
          item.classList.add("opacity-modified");
        });
    }
  });
});
