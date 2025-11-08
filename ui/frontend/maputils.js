const map1 = ['#fff7ec', '#fee8c8', '#fdd49e', '#fdbb84', '#fc8d59', '#ef6548', '#d7301f', '#b30000', '#7f0000'];

//const map2 = ['#a50026','#d73027','#f46d43','#fdae61','#fee090','#ffffbf','#e0f3f8','#abd9e9','#74add1','#4575b4','#313695'];

const nbclasses = map1.length
const maxValue = 3000
const minValue = 500
const classRange = (maxValue - minValue) / (nbclasses - 2)

export function getColorFromPrice(v) {
    
    if (v < minValue) return map1[0];
    if (v > maxValue) return map1[nbclasses-1];

    const iv = Math.floor((v - minValue) / classRange);
    const c = map1[iv];
    return c;
}
