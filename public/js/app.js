console.log("KVolt Static File Test: JS Loaded!");
document.addEventListener("DOMContentLoaded", () => {
    const p = document.createElement("p");
    p.innerText = "JS successfully executed!";
    p.style.color = "green";
    document.body.appendChild(p);
});
