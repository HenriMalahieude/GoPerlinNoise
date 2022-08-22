document.getElementById("grid").value = 10
document.getElementById("resolution").value = 15

let output = document.getElementById("p1");
let thread_out = document.getElementById("p2")
let but = document.getElementById("Refresh");

but.onclick = function(){
    let grid = document.getElementById("grid").value
    let res = document.getElementById("resolution").value
    
    output.innerHTML = "Grid of " + grid + " x " + grid + " squares @ a resolution of " + res + " x " + res + " per square";
    thread_out.innerHTML = "Gradient Threads: " + (grid) + ", Generation Threads: " + (grid * res);
    generate(grid, res);
}

function generate(g, r){
    let canv = document.getElementById("window")
    let ctx = canv.getContext("2d");
    ctx.clearRect(0, 0, canv.clientWidth, canv.clientHeight);

    but.innerHTML = genPerlin(g, r)
}