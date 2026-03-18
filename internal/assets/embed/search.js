(async () => {
  const resultsContainer = document.getElementById('search-results');
  const resultsEl = resultsContainer.querySelector(':scope > div');
  const input = document.getElementById('navbar-search-input');

  const pagefind = await import(resultsContainer.dataset.pagefind);
  await pagefind.options({
    baseUrl: '/',
  });

  function closeSearch() {
    resultsEl.innerHTML = '';
    resultsContainer.classList.remove('is-active');
  }

  function renderResults(search) {
    resultsEl.innerHTML = '';
    if (!search || search.results.length === 0) {
      if (input.value.trim().length >= 2) {
        const empty = document.createElement('div');
        empty.className = 'navbar-search-empty';
        empty.textContent = 'No results found.';
        resultsEl.appendChild(empty);
      }
      resultsContainer.classList.toggle('is-active', input.value.trim().length >= 2);
      return;
    }
    Promise.all(search.results.slice(0, 8).map(r => r.data())).then(results => {
      resultsEl.innerHTML = '';
      results.forEach(r => {
        const a = document.createElement('a');
        a.className = 'navbar-search-result';
        a.href = r.url;
        const title = document.createElement('strong');
        title.textContent = r.meta?.title || r.url;
        const excerpt = document.createElement('span');
        excerpt.innerHTML = r.excerpt;
        a.append(title, excerpt);
        resultsEl.appendChild(a);
      });
      resultsContainer.classList.add('is-active');
    });
  }

  input.addEventListener('input', async () => {
    const query = input.value.trim();
    if (query.length < 2) {
      resultsEl.innerHTML = '';
      resultsContainer.classList.remove('is-active');
      return;
    }
    const search = await pagefind.debouncedSearch(query, {}, 200);
    if (search) {
      renderResults(search);
    }
  });

  input.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') {
      input.value = '';
      closeSearch();
      input.blur();
    }
  });

  input.addEventListener('focus', () => {
    const burger = document.querySelector('.navbar-burger');
    if (burger && getComputedStyle(burger).display !== 'none') {
      input.scrollIntoView({
        behavior: 'smooth',
        block: 'start',
      });
    }
  });

  input.addEventListener('blur', () => {
    setTimeout(() => {
      if (!document.activeElement?.closest('#search-results')) {
        closeSearch();
      }
    }, 150);
  });

  resultsContainer.addEventListener('click', (e) => {
    if (e.target.closest('a')) {
      input.value = '';
      closeSearch();
    }
  });

  window.addEventListener('pageshow', (e) => {
    if (e.persisted) {
      input.value = '';
      closeSearch();
    }
  });
})();
