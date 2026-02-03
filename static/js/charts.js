/**
 * charts.js - Chart.js helper functions for Kotomi Analytics
 */

/**
 * Create a line chart for time series data
 * @param {string} canvasId - ID of the canvas element
 * @param {string} label - Chart label
 * @param {object} data - Data object with labels and values arrays
 */
function createLineChart(canvasId, label, data) {
    const ctx = document.getElementById(canvasId);
    if (!ctx) {
        console.error('Canvas element not found:', canvasId);
        return;
    }

    new Chart(ctx, {
        type: 'line',
        data: {
            labels: data.labels,
            datasets: [{
                label: label,
                data: data.values,
                borderColor: 'rgb(75, 192, 192)',
                backgroundColor: 'rgba(75, 192, 192, 0.2)',
                tension: 0.3,
                fill: true
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: true,
            plugins: {
                legend: {
                    display: true,
                    position: 'top'
                },
                tooltip: {
                    mode: 'index',
                    intersect: false
                }
            },
            scales: {
                y: {
                    beginAtZero: true,
                    ticks: {
                        precision: 0
                    }
                }
            }
        }
    });
}

/**
 * Create a pie chart for distribution data
 * @param {string} canvasId - ID of the canvas element
 * @param {string} label - Chart label
 * @param {object} data - Data object with labels and values arrays
 */
function createPieChart(canvasId, label, data) {
    const ctx = document.getElementById(canvasId);
    if (!ctx) {
        console.error('Canvas element not found:', canvasId);
        return;
    }

    // Generate colors for pie chart
    const colors = generateColors(data.labels.length);

    new Chart(ctx, {
        type: 'pie',
        data: {
            labels: data.labels,
            datasets: [{
                label: label,
                data: data.values,
                backgroundColor: colors.background,
                borderColor: colors.border,
                borderWidth: 1
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: true,
            plugins: {
                legend: {
                    display: true,
                    position: 'right'
                },
                tooltip: {
                    callbacks: {
                        label: function(context) {
                            const label = context.label || '';
                            const value = context.parsed || 0;
                            const total = context.dataset.data.reduce((a, b) => a + b, 0);
                            const percentage = ((value / total) * 100).toFixed(1);
                            return `${label}: ${value} (${percentage}%)`;
                        }
                    }
                }
            }
        }
    });
}

/**
 * Create a bar chart for comparison data
 * @param {string} canvasId - ID of the canvas element
 * @param {string} label - Chart label
 * @param {object} data - Data object with labels and values arrays
 */
function createBarChart(canvasId, label, data) {
    const ctx = document.getElementById(canvasId);
    if (!ctx) {
        console.error('Canvas element not found:', canvasId);
        return;
    }

    new Chart(ctx, {
        type: 'bar',
        data: {
            labels: data.labels,
            datasets: [{
                label: label,
                data: data.values,
                backgroundColor: 'rgba(54, 162, 235, 0.5)',
                borderColor: 'rgb(54, 162, 235)',
                borderWidth: 1
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: true,
            plugins: {
                legend: {
                    display: true,
                    position: 'top'
                }
            },
            scales: {
                y: {
                    beginAtZero: true,
                    ticks: {
                        precision: 0
                    }
                }
            }
        }
    });
}

/**
 * Generate colors for charts
 * @param {number} count - Number of colors to generate
 * @returns {object} - Object with background and border color arrays
 */
function generateColors(count) {
    const baseColors = [
        'rgb(255, 99, 132)',
        'rgb(54, 162, 235)',
        'rgb(255, 206, 86)',
        'rgb(75, 192, 192)',
        'rgb(153, 102, 255)',
        'rgb(255, 159, 64)',
        'rgb(201, 203, 207)',
        'rgb(255, 99, 255)',
        'rgb(99, 255, 132)',
        'rgb(132, 99, 255)'
    ];

    const background = [];
    const border = [];

    for (let i = 0; i < count; i++) {
        const color = baseColors[i % baseColors.length];
        border.push(color);
        // Convert to rgba with 0.5 alpha for background
        background.push(color.replace('rgb', 'rgba').replace(')', ', 0.5)'));
    }

    return { background, border };
}

/**
 * Update chart data dynamically
 * @param {Chart} chart - Chart.js instance
 * @param {object} newData - New data object with labels and values
 */
function updateChartData(chart, newData) {
    if (!chart) {
        console.error('Chart instance not provided');
        return;
    }

    chart.data.labels = newData.labels;
    chart.data.datasets[0].data = newData.values;
    chart.update();
}
