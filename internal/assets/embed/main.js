document.addEventListener('DOMContentLoaded', () => {
  anchors.add('.content h1, .content h2, .content h3, .content h4, .content h5, .content h6');
  document.querySelectorAll('.content figure:not(.image) > img').forEach(img => {
    const figcaption = img.closest('figure')?.querySelector('figcaption');
    if (figcaption) {
      img.dataset.title = figcaption.textContent.trim();
    }
  });
  GLightbox({
    touchNavigation: true,
    selector: '.content figure:not(.image) > img',
  });
  document.querySelectorAll('.navbar-burger').forEach(element => {
    element.addEventListener('click', () => {
      element.classList.toggle('is-active');
      document.getElementById(element.dataset.target).classList.toggle('is-active');
    });
  });
  document.querySelectorAll('.content table').forEach(table => {
    const wrapper = document.createElement('div');
    wrapper.className = 'table-container';
    table.parentNode.insertBefore(wrapper, table);
    wrapper.appendChild(table);
  });
});
