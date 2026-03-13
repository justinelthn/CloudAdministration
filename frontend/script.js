const API_BASE = "";

// ── Autocomplete for actor & director fields ──
function setupAutocomplete(inputId, datalistId, field) {
    const input = document.getElementById(inputId);
    const datalist = document.getElementById(datalistId);
    let debounce = null;

    input.addEventListener("input", () => {
        clearTimeout(debounce);
        const q = input.value.trim();
        if (q.length < 2) {
            datalist.innerHTML = "";
            return;
        }
        debounce = setTimeout(() => {
            fetch(`${API_BASE}/suggest?field=${field}&query=${encodeURIComponent(q)}`)
                .then((res) => res.json())
                .then((items) => {
                    datalist.innerHTML = "";
                    items.forEach((item) => {
                        const opt = document.createElement("option");
                        opt.value = item;
                        datalist.appendChild(opt);
                    });
                })
                .catch((err) => console.warn("Autocomplete error:", err));
        }, 300);
    });
}

setupAutocomplete("actor", "actorList", "actors");
setupAutocomplete("director", "directorList", "director");

// ── Search ──
document.getElementById("searchBtn").addEventListener("click", () => {
    const actor = document.getElementById("actor").value;
    const director = document.getElementById("director").value;
    const genres = Array.from(
        document.getElementById("genre").selectedOptions
    ).map((o) => o.value);
    const languages = Array.from(
        document.getElementById("language").selectedOptions
    ).map((o) => o.value);
    const min_rating = document.getElementById("min_rating").value;

    const errorEl = document.getElementById("error-msg");
    const countEl = document.getElementById("result-count");
    errorEl.classList.add("hidden");
    countEl.classList.add("hidden");

    let url = `${API_BASE}/movies?min_rating=${min_rating}`;
    if (actor) url += `&actors=${encodeURIComponent(actor)}`;
    if (director) url += `&directors=${encodeURIComponent(director)}`;
    genres.forEach((g) => (url += `&genres=${encodeURIComponent(g)}`));
    languages.forEach((l) => (url += `&languages=${encodeURIComponent(l)}`));

    fetch(url)
        .then((res) => {
            if (!res.ok) throw new Error(`Erreur serveur (${res.status})`);
            return res.json();
        })
        .then((data) => {
            const tbody = document.querySelector("#results tbody");
            tbody.innerHTML = "";

            if (!data || data.length === 0) {
                tbody.innerHTML =
                    '<tr><td colspan="7" style="text-align:center;color:#888;">Aucun résultat trouvé</td></tr>';
                countEl.textContent = "0 résultat(s)";
            } else {
                countEl.textContent = `${data.length} résultat(s)`;
                data.forEach((m) => {
                    const tr = document.createElement("tr");
                    tr.innerHTML = `<td>${esc(m.title)}</td><td>${esc(m.director)}</td><td>${esc((m.actors || []).join(", "))}</td><td>${esc((m.genres || []).join(", "))}</td><td>${esc(m.original_language)}</td><td>${m.average_rating}</td><td>${m.runtime}</td>`;
                    tbody.appendChild(tr);
                });
            }
            countEl.classList.remove("hidden");
        })
        .catch((err) => {
            errorEl.textContent = `Erreur : ${err.message}`;
            errorEl.classList.remove("hidden");
        });
});

// Simple HTML escape
function esc(str) {
    if (!str) return "";
    const d = document.createElement("div");
    d.textContent = str;
    return d.innerHTML;
}
