function initializeMap() {
  var mapElement = document.getElementById("map");
  if (mapElement) {
    var lat = parseFloat(mapElement.getAttribute("data-latitude"));
    var lng = parseFloat(mapElement.getAttribute("data-longitude"));

    map = L.map("map").setView([lat, lng], 13);
    marker = L.marker([lat, lng], { draggable: true }).addTo(map);

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
    L.Control.geocoder().addTo(map);
    return map;
  }
}

document.addEventListener("click", function (event) {
  var panel = document.getElementById("side-panel");
  if (!panel.contains(event.target) && !event.target.matches(".media-item")) {
    panel.classList.remove("open");
    document.getElementById("media-list").classList.remove("side-panel-open");
    document.getElementById("panel-content").innerHTML = ""; // Clear the content
  }

  setTimeout(() => {
    if (document.getElementById("panel-content").innerHTML == "") {
      document
        .querySelectorAll(".media-item.active")
        .forEach((m) => m.classList.remove("active"));
    }
  }, 500);
});

function updateActiveMediaItem(target) {
  var activeElements = document.querySelectorAll(".media-item.active");
  activeElements.forEach(function (el) {
    el.classList.remove("active");
  });
  target.classList.add("active");

  // Scroll to the active element
  target.scrollIntoView({
    behavior: "smooth",
    block: "center",
  });

  // Load media details
  htmx.trigger(target, "click");
}

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
  console.log(evt);
  const trigger = evt.detail.xhr.getResponseHeader("HX-Trigger");
  if (trigger === "mediaUpdated") {
    closePanel();
  }

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

// function closePanel() {
//   var panel = document.getElementById("side-panel");
//   panel.classList.remove("open");
//   document.getElementById("media-list").classList.remove("side-panel-open");
//   document.getElementById("panel-content").innerHTML = ""; // Clear the content
// }
//
// let onlyChangeClick = false;
//
// document.getElementById("only-change").addEventListener("click", () => {
//   const items = document.querySelectorAll(".media-item");
//   items.forEach((i) => {
//     if (!i.classList.contains("modified") && !onlyChangeClick) {
//       i.style.opacity = 0.2;
//     } else {
//       i.style.opacity = 1;
//     }
//   });
//   onlyChangeClick = !onlyChangeClick;
// });
