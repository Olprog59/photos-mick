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

  console.log(evt.detail);
  const failed = evt.detail.failed;

  if (failed) {
    popover.className = "";
    footer.classList.add("active");
    popover.querySelector("p").innerText = evt.detail.xhr.responseText;
    popover.classList.add("error");
  } else {
    const header = evt.detail.xhr.getResponseHeader("Message");

    if (header) {
      try {
        // Décodez le message Base64
        const decodedBytes = atob(header);
        // Convertissez les bytes décodés en une chaîne UTF-8
        const decodedMessage = new TextDecoder("utf-8").decode(
          new Uint8Array([...decodedBytes].map((c) => c.charCodeAt(0))),
        );

        console.log("Message reçu (Base64):", header);
        console.log("Message après atob:", decodedBytes);
        console.log("Message final décodé:", decodedMessage);

        popover.className = "";
        footer.classList.add("active");
        popover.querySelector("p").innerText = decodedMessage;
        popover.classList.add("success");
      } catch (error) {
        console.error("Erreur lors du décodage:", error);
        popover.querySelector("p").innerText =
          "Erreur lors du décodage du message";
        popover.classList.add("error");
      }
    }
  }
});

// document.addEventListener("htmx:configRequest", function (evt) {
//   evt.detail.headers["HX-Trigger"] = "mediaUpdated";
// });
//
// document.addEventListener("mediaUpdated", function () {
//   htmx.ajax("GET", "/api/medias", "#media-list");
// });
