document.addEventListener("DOMContentLoaded", function () {
  const linkCopiers = document.querySelectorAll("button.link-copier")
  const images = document.querySelectorAll(".article-post img, .post-content-section img, .images-container img");

  linkCopiers.forEach(copyArticleLink);
  images.forEach(openImageInViewer);
});

function openImageInViewer(img) {
  img.style.cursor = "zoom-in";
  img.addEventListener("click", handleImageClicked.bind(this, img));
}

function handleImageClicked(img) {
  const imageDialog = document.createElement("dialog");
  const imageDialogHeading = document.createElement("p");
  const imageContainerDiv = document.createElement("div");
  const image = document.createElement("img");
  let imageCaption;

  imageDialog.classList.add("img-viewer");

  imageDialogHeading.classList.add("small");
  imageDialogHeading.textContent = "Hit `^Esc` or tap outside the image to close."
  imageDialog.appendChild(imageDialogHeading);

  imageContainerDiv.classList.add("img-container");

  image.src = img.src;
  image.alt = img.alt;

  imageContainerDiv.appendChild(image);
  imageDialog.appendChild(imageContainerDiv);

  /* Set image caption.  */

  let currentCaption = img.parentNode.querySelector(".caption p");
  if (currentCaption != null) { /* When project or article post image.  */
    imageCaption = currentCaption.cloneNode(true); /* Clone the whole caption */
  } else {
    currentCaption = img.parentNode.parentNode.querySelector("small");
    if (currentCaption != null) { /* When archive article cover image.  */
      imageCaption = document.createElement("p");
      imageCaption.textContent = currentCaption.textContent;
    } else { /* When project cover image.  */
      let element = img;
      let found = true;
      while ((element = element.parentNode)) { /* Stops at either '.project-tile' or '.info-article'.  */
        if (element.nodeName.toLowerCase() === "main") { /* When, for instance, there is no caption at all.  */
          found = false;
          break;
        }

        if (element.classList.contains("project-tile") || element.classList.contains("info-article")) {
          break
        }
      }

      if (found) {
        const title = element.querySelector("p.name, h1.name");
        imageCaption = document.createElement("p");
        imageCaption.textContent = title.textContent;
      }
    }
  }

  if (imageCaption != null) {
    imageCaption.classList.add("caption");
    imageDialog.appendChild(imageCaption);
  }

  document.body.appendChild(imageDialog);
  imageDialog.showModal();

  /* Close dialog when clicking outside the image.  */
  imageDialog.addEventListener("click", (e) => {
    const rect = imageDialog.getBoundingClientRect();
    const isOutside =
      e.clientX < rect.left || e.clientX > rect.right ||
      e.clientY < rect.top || e.clientY > rect.bottom;
    if (isOutside) {
      imageDialog.close();
    }
  });

  /* Close dialog when hitting '^Esc'.  */
  const escHandler = (e) => {
    if (e.key === "Escape") {
      imageDialog.close();
    }
  };

  document.addEventListener("keydown", escHandler);

  imageDialog.addEventListener("close", () => {
    document.removeEventListener("keydown", escHandler);
    imageDialog.remove();
  });
}

function copyArticleLink(button) {
  let isCopied = false;
  button.onclick = () => {
    if (isCopied) {
      return;
    }

    const dummy = document.createElement('input');

    document.body.appendChild(dummy);
    dummy.value = window.location.href;
    dummy.select();

    try {
      document.execCommand('copy');
      isCopied = true;
      button.classList.add("copied");
      button.textContent = "Copied!";
      setTimeout(() => {
        button.textContent = "Copy link";
        button.classList.remove("copied");
        isCopied = false;
      }, 5000);
    } catch (error) {
      console.error(error);
    } finally {
      document.body.removeChild(dummy);
    }
  };
}

function toggleNavigationSidebar() {
  const menu = document.getElementById("navigation")
  const backdrop = document.getElementById("backdrop")
  menu.classList.toggle("show")
  backdrop.classList.toggle("show")
}

function searchArticles(e) {
  const query = e.value.trim();
  let newURL = window.location.pathname;

  if ("" !== query) {
    newURL += "?search=" + query;
  }

  e.setAttribute("hx-get", newURL);
  e.setAttribute("hx-push-url", newURL);
  window.history.replaceState({}, "", newURL);
}

function setArchiveTopic(e) {
  document.querySelector("h3.topic-and-date").classList.remove("hide");
  document.querySelector("div.selected-tag-div").classList.add("hide");
  document.querySelector("span.selected-topic").textContent = e.textContent;
  document.querySelector("span.selected-date").textContent = "Any date";

  for (const topicItem of document.getElementById("topics-list").children) {
    topicItem.classList.remove("selected");
  }

  for (const tagItem of document.getElementById("tags-list").children) {
    tagItem.classList.remove("selected");
  }

  const currentTopicItem = e.parentNode;
  currentTopicItem.classList.add("selected");

  const baseTopicURL = e.getAttribute("hx-get");
  const publicationsList = document.getElementById("publications-list").children;

  for (const publicationItem of publicationsList) {
    publicationItem.classList.remove("selected");

    let newTopicURL = baseTopicURL;
    const anchor = publicationItem.firstChild;
    const hxGetValue = anchor.getAttribute("hx-get").split("/");
    const month = hxGetValue.at(-1);
    const year = hxGetValue.at(-2);

    newTopicURL += "/" + year + "/" + month;
    anchor.setAttribute("href", newTopicURL);
    anchor.setAttribute("hx-get", newTopicURL);
    anchor.setAttribute("hx-push-url", "true");
  }

  const searchbar = document.getElementById("searchbar");

  searchbar.value = "";

  searchbar.setAttribute("hx-get", baseTopicURL);
  searchbar.setAttribute("hx-push-url", baseTopicURL);
  window.history.replaceState({}, "", baseTopicURL);
  window.scrollTo({top: 0, behavior: "smooth"});

  htmx.process(document.body);
}

function setArchivePublicationDate(e) {
  document.querySelector("h3.topic-and-date").classList.remove("hide");
  document.querySelector("div.selected-tag-div").classList.add("hide");
  document.querySelector("span.selected-date").textContent = e.textContent;

  for (const publicationItem of document.getElementById("publications-list").children) {
    publicationItem.classList.remove("selected");
  }

  for (const tagItem of document.getElementById("tags-list").children) {
    tagItem.classList.remove("selected");
  }

  const currentPublicationItem = e.parentNode;
  const currentURL = e.getAttribute("href");

  currentPublicationItem.classList.add("selected");

  const searchbar = document.getElementById("searchbar");

  searchbar.value = "";
  searchbar.setAttribute("hx-get", currentURL);
  searchbar.setAttribute("hx-push-url", currentURL);

  window.history.replaceState({}, "", currentURL);
  window.scrollTo({top: 0, behavior: "smooth"});

  htmx.process(document.body);
}

function setArchiveTag(e) {
  document.querySelector("h3.topic-and-date").classList.add("hide");
  document.querySelector("div.selected-tag-div").classList.remove("hide");
  document.querySelector(".selected-tag").textContent = e.textContent;

  for (const topicItem of document.getElementById("topics-list").children) {
    topicItem.classList.remove("selected");
  }

  for (const tagItem of document.getElementById("tags-list").children) {
    tagItem.classList.remove("selected");
  }

  const currentTagItem = e.parentNode;
  currentTagItem.classList.add("selected");

  const baseTagURL = e.getAttribute("hx-get");

  for (const publicationItem of document.getElementById("publications-list").children) {
    publicationItem.classList.remove("selected");

    let publicationURL = "/archive/any";
    const anchor = publicationItem.firstChild;
    const hxGetValue = anchor.getAttribute("hx-get").split("/");
    const month = hxGetValue.at(-1);
    const year = hxGetValue.at(-2);

    publicationURL += "/" + year + "/" + month;
    anchor.setAttribute("href", publicationURL);
    anchor.setAttribute("hx-get", publicationURL);
    anchor.setAttribute("hx-push-url", "true");
  }

  const searchbar = document.getElementById("searchbar");

  searchbar.value = "";

  searchbar.setAttribute("hx-get", baseTagURL);
  searchbar.setAttribute("hx-push-url", baseTagURL);
  window.history.replaceState({}, "", baseTagURL);
  window.scrollTo({top: 0, behavior: "smooth"});

  htmx.process(document.body);
}
