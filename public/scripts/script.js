function toggleNavigationSidebar() {
  const menu = document.getElementById("navigation")
  const backdrop = document.getElementById("backdrop")
  menu.classList.toggle("show")
  backdrop.classList.toggle("show")
}

async function copyLinkToClipboard() {
  await navigator.clipboard.writeText(window.location.href);
}
