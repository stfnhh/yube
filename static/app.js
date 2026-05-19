const themeIcons = {
  dark: '<i class="icon icon-moon" aria-hidden="true"></i>',
  light: '<i class="icon icon-sun" aria-hidden="true"></i>',
};

initThemeToggle();
initMobileSidebar();
initSearchControls();
initVideoHoverColors();
initWatchCountUpdates();
initUnsubscribeButtons();
initUploadZones();

function initThemeToggle() {
  const body = document.body;
  const toggle = document.getElementById("themeToggle");

  if (!toggle) {
    return;
  }

  setTheme(localStorage.getItem("theme") === "dark");

  toggle.addEventListener("click", () => {
    setTheme(!body.classList.contains("dark"));
  });

  function setTheme(isDark) {
    body.classList.toggle("dark", isDark);
    toggle.innerHTML = isDark ? themeIcons.dark : themeIcons.light;
    localStorage.setItem("theme", isDark ? "dark" : "light");
  }
}

function initMobileSidebar() {
  const menuButton = document.getElementById("mobileMenuBtn");
  const overlay = document.getElementById("mobileOverlay");
  const sidebar = document.querySelector(".sidebar");

  if (!menuButton || !overlay || !sidebar) {
    return;
  }

  const close = () => {
    sidebar.classList.remove("mobile-open");
    overlay.classList.remove("show");
  };

  menuButton.addEventListener("click", () => {
    sidebar.classList.add("mobile-open");
    overlay.classList.add("show");
  });

  overlay.addEventListener("click", close);
  window.addEventListener("resize", () => {
    if (window.innerWidth > 1100) {
      close();
    }
  });
}

function initSearchControls() {
  document.querySelectorAll(".search").forEach((form) => {
    const input = form.querySelector('input[type="search"]');
    const clearButton = form.querySelector(".search-clear");

    if (!input || !clearButton) {
      return;
    }

    const syncClearButton = () => {
      clearButton.classList.toggle("visible", input.value.length > 0);
    };

    clearButton.addEventListener("click", () => {
      input.value = "";
      input.focus();
      syncClearButton();
    });

    input.addEventListener("input", syncClearButton);
    syncClearButton();
  });
}

function initVideoHoverColors() {
  document.querySelectorAll(".video-card").forEach((card) => {
    const image = card.querySelector(".thumbnail img");

    if (!image) {
      return;
    }

    const applyColor = () => {
      const color = getImageAverageColor(image);

      if (!color) {
        return;
      }

      setHoverColor(card, color, "--video-hover-bg", 0.18);
      setHoverColor(card, color, "--video-hover-bg-dark", 0.24);
    };

    if (image.complete && image.naturalWidth > 0) {
      applyColor();
      return;
    }

    image.addEventListener("load", applyColor, { once: true });
  });
}

function setHoverColor(element, color, property, alpha) {
  element.style.setProperty(
    property,
    `rgba(${color.r}, ${color.g}, ${color.b}, ${alpha})`,
  );
}

function initWatchCountUpdates() {
  document.querySelectorAll(".video-card-link").forEach((link) => {
    link.addEventListener("click", () => {
      if (link.dataset.watched === "true") {
        return;
      }

      const count = channelCountElement(link.dataset.channelId);
      if (count) {
        count.textContent = String(
          Math.max(0, Number.parseInt(count.textContent.trim(), 10) - 1),
        );
      }

      markVideoCardViewed(link);
      link.dataset.watched = "true";
    });
  });
}

function markVideoCardViewed(link) {
  const meta = link.querySelector(".video-meta");

  if (!meta || meta.querySelector(".watched-indicator")) {
    return;
  }

  const dot = document.createElement("span");
  dot.className = "video-meta-dot";
  dot.textContent = "•";

  const indicator = document.createElement("span");
  indicator.className = "watched-indicator";
  indicator.setAttribute("aria-label", "Viewed");
  indicator.setAttribute("title", "Viewed");
  indicator.innerHTML = '<i class="icon icon-check" aria-hidden="true"></i>';

  meta.append(dot, indicator);
}

function channelCountElement(channelID) {
  if (!channelID) {
    return null;
  }

  return document.querySelector(
    `.sub-item[data-channel-id="${CSS.escape(channelID)}"] .sub-count`,
  );
}

function initUnsubscribeButtons() {
  document.querySelectorAll(".unsubscribe-btn[data-delete-url]").forEach((button) => {
    button.addEventListener("click", () => unsubscribe(button));
  });
}

async function unsubscribe(button) {
  const channelTitle = button.dataset.channelTitle || "this channel";

  if (!window.confirm(`Unsubscribe from ${channelTitle}?`)) {
    return;
  }

  button.disabled = true;

  const response = await fetch(button.dataset.deleteUrl, {
    method: "DELETE",
    headers: {
      "Accept": "text/html",
    },
  });

  if (response.redirected) {
    window.location.assign(response.url);
    return;
  }

  if (response.ok) {
    window.location.assign("/channels");
    return;
  }

  button.disabled = false;
}

function initUploadZones() {
  document.querySelectorAll(".upload-zone").forEach((zone) => {
    const input = zone.querySelector('input[type="file"]');
    const target = zone.querySelector(".upload-drop-target");
    const fileName = zone.querySelector("[data-file-name]");

    if (!input || !target) {
      return;
    }

    input.addEventListener("change", () => {
      updateSelectedFileName(input, fileName);
    });

    zone.addEventListener("submit", (event) => {
      if (input.files.length > 0) {
        return;
      }

      event.preventDefault();
      input.click();
    });

    target.addEventListener("keydown", (event) => {
      if (event.key !== "Enter" && event.key !== " ") {
        return;
      }

      event.preventDefault();
      input.click();
    });

    ["dragenter", "dragover"].forEach((eventName) => {
      target.addEventListener(eventName, (event) => {
        event.preventDefault();
        zone.classList.add("drag-over");
      });
    });

    ["dragleave", "drop"].forEach((eventName) => {
      target.addEventListener(eventName, () => {
        zone.classList.remove("drag-over");
      });
    });

    target.addEventListener("drop", (event) => {
      event.preventDefault();

      if (!event.dataTransfer || event.dataTransfer.files.length === 0) {
        return;
      }

      input.files = event.dataTransfer.files;
      updateSelectedFileName(input, fileName);
    });
  });
}

function updateSelectedFileName(input, fileName) {
  if (!fileName) {
    return;
  }

  fileName.textContent = input.files.length > 0
    ? input.files[0].name
    : "No file selected";
}

function getImageAverageColor(image) {
  const canvas = document.createElement("canvas");
  const width = 48;
  const height = 27;

  canvas.width = width;
  canvas.height = height;

  const context = canvas.getContext("2d", {
    willReadFrequently: true,
  });

  if (!context) {
    return null;
  }

  try {
    context.drawImage(image, 0, 0, width, height);
    return strongestImageColor(
      context.getImageData(0, 0, width, height).data,
      width,
      height,
    );
  } catch {
    return null;
  }
}

function strongestImageColor(data, width, height) {
  const colorfulBuckets = new Map();
  const fallbackBuckets = new Map();

  for (let i = 0; i < data.length; i += 4) {
    const color = imagePixelColor(data, i, width, height);

    if (!color) {
      continue;
    }

    addColorToBucket(fallbackBuckets, color.fallbackKey, color, color.fallbackWeight);

    if (color.hsl.s >= 0.2) {
      addColorToBucket(colorfulBuckets, color.key, color, color.colorWeight);
    }
  }

  const winner = strongestBucket(colorfulBuckets) || strongestBucket(fallbackBuckets);
  if (!winner) {
    return null;
  }

  return boostColor({
    r: Math.round(winner.r / winner.weight),
    g: Math.round(winner.g / winner.weight),
    b: Math.round(winner.b / winner.weight),
  });
}

function imagePixelColor(data, index, width, height) {
  const red = data[index];
  const green = data[index + 1];
  const blue = data[index + 2];
  const alpha = data[index + 3];

  if (alpha < 200) {
    return null;
  }

  const hsl = rgbToHsl(red, green, blue);
  const lightness = hsl.l * 255;

  if (lightness < 18 || lightness > 238) {
    return null;
  }

  const pixel = index / 4;
  const x = pixel % width;
  const y = Math.floor(pixel / width);
  const edge =
    x < width * 0.18 ||
    x > width * 0.82 ||
    y < height * 0.18 ||
    y > height * 0.82;
  const edgeWeight = edge ? 2.4 : 1;

  return {
    r: red,
    g: green,
    b: blue,
    hsl,
    key: [
      Math.round(hsl.h / 18),
      Math.round(hsl.s * 5),
      Math.round(hsl.l * 5),
    ].join(","),
    fallbackKey: [
      Math.round(red / 32),
      Math.round(green / 32),
      Math.round(blue / 32),
    ].join(","),
    colorWeight: edgeWeight * (0.4 + hsl.s * hsl.s * 5),
    fallbackWeight: edgeWeight * (0.8 + hsl.s),
  };
}

function addColorToBucket(buckets, key, color, weight) {
  const bucket = buckets.get(key) || {
    r: 0,
    g: 0,
    b: 0,
    weight: 0,
  };

  bucket.r += color.r * weight;
  bucket.g += color.g * weight;
  bucket.b += color.b * weight;
  bucket.weight += weight;

  buckets.set(key, bucket);
}

function strongestBucket(buckets) {
  let winner = null;

  buckets.forEach((bucket) => {
    if (!winner || bucket.weight > winner.weight) {
      winner = bucket;
    }
  });

  return winner;
}

function boostColor(color) {
  const hsl = rgbToHsl(color.r, color.g, color.b);

  if (hsl.s < 0.12) {
    return color;
  }

  return hslToRgb(
    hsl.h,
    Math.min(0.78, Math.max(0.42, hsl.s * 1.45)),
    Math.min(0.62, Math.max(0.34, hsl.l)),
  );
}

function rgbToHsl(red, green, blue) {
  const r = red / 255;
  const g = green / 255;
  const b = blue / 255;
  const max = Math.max(r, g, b);
  const min = Math.min(r, g, b);
  const l = (max + min) / 2;

  if (max === min) {
    return { h: 0, s: 0, l };
  }

  const d = max - min;
  const s = l > 0.5
    ? d / (2 - max - min)
    : d / (max + min);
  let h;

  if (max === r) {
    h = (g - b) / d + (g < b ? 6 : 0);
  } else if (max === g) {
    h = (b - r) / d + 2;
  } else {
    h = (r - g) / d + 4;
  }

  return {
    h: h * 60,
    s,
    l,
  };
}

function hslToRgb(hue, saturation, lightness) {
  const c = (1 - Math.abs(2 * lightness - 1)) * saturation;
  const x = c * (1 - Math.abs(((hue / 60) % 2) - 1));
  const m = lightness - c / 2;
  let r = 0;
  let g = 0;
  let b = 0;

  if (hue < 60) {
    r = c;
    g = x;
  } else if (hue < 120) {
    r = x;
    g = c;
  } else if (hue < 180) {
    g = c;
    b = x;
  } else if (hue < 240) {
    g = x;
    b = c;
  } else if (hue < 300) {
    r = x;
    b = c;
  } else {
    r = c;
    b = x;
  }

  return {
    r: Math.round((r + m) * 255),
    g: Math.round((g + m) * 255),
    b: Math.round((b + m) * 255),
  };
}
