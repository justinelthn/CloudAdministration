document.getElementById("searchBtn").addEventListener("click", () => {
    const actor = document.getElementById("actor").value
    const director = document.getElementById("director").value
    const genres = Array.from(document.getElementById("genre").selectedOptions).map(o => o.value)
    const languages = Array.from(document.getElementById("language").selectedOptions).map(o => o.value)
    const min_rating = document.getElementById("min_rating").value

    let url = `/movies?min_rating=${min_rating}`
    if(actor) url += `&actors=${encodeURIComponent(actor)}`
    if(director) url += `&directors=${encodeURIComponent(director)}`
    genres.forEach(g => url += `&genres=${encodeURIComponent(g)}`)
    languages.forEach(l => url += `&languages=${encodeURIComponent(l)}`)

    fetch(url)
    .then(res => res.json())
    .then(data => {
        const tbody = document.querySelector("#results tbody")
        tbody.innerHTML = ""
        data.forEach(m => {
            const tr = document.createElement("tr")
            tr.innerHTML = `<td>${m.title}</td><td>${m.director}</td><td>${m.actors.join(", ")}</td><td>${m.genres.join(", ")}</td><td>${m.original_language}</td><td>${m.average_rating}</td><td>${m.runtime}</td>`
            tbody.appendChild(tr)
        })
    })
})
