document.getElementById("searchBtn").addEventListener("click", () => {
    const director = document.getElementById("director").value
    const actor = document.getElementById("actor").value
    const genre = document.getElementById("genre").value
    const language = document.getElementById("language").value
    const min_rating = document.getElementById("min_rating").value

    let url = "/movies?"
    if (director) url += "directors=" + encodeURIComponent(director) + "&"
    if (actor) url += "actors=" + encodeURIComponent(actor) + "&"
    if (genre) url += "genres=" + encodeURIComponent(genre) + "&"
    if (language) url += "languages=" + encodeURIComponent(language) + "&"
    if (min_rating) url += "min_rating=" + min_rating

    fetch(url)
        .then(r => r.json())
        .then(data => {
            const tbody = document.querySelector("#results tbody")
            tbody.innerHTML = ""
            document.getElementById("info").textContent = data.length + " résultat(s)"
            data.forEach(m => {
                const tr = document.createElement("tr")
                tr.innerHTML = `<td>${m.title}</td><td>${m.director}</td><td>${(m.actors || []).join(", ")}</td><td>${(m.genres || []).join(", ")}</td><td>${m.original_language}</td><td>${m.average_rating}</td>`
                tbody.appendChild(tr)
            })
        })
        .catch(e => {
            document.getElementById("info").textContent = "Erreur : " + e.message
        })
})
