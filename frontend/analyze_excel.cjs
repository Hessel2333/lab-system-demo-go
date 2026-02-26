const XLSX = require('xlsx');
const fs = require('fs');
const path = require('path');

const filePath = path.join(__dirname, '../2026年1月易派客采购明细.xls');
const outputPath = path.join(__dirname, '../易派客明细结构分析.md');

try {
    const wb = XLSX.readFile(filePath);
    let output = '# 易派客采购明细 Excel 结构分析\n\n';

    wb.SheetNames.forEach(sheetName => {
        output += `## 工作表: ${sheetName}\n\n`;
        const ws = wb.Sheets[sheetName];

        // 提取带 header: 1 的二维数组格式，方便看到绝对的行列关系
        const data = XLSX.utils.sheet_to_json(ws, { header: 1 });

        output += `总行数: ${data.length}\n\n`;

        if (data.length > 0) {
            // 探索推断表头所在的行（通常非空字段最多的前几行就是表头）
            let headerRowIndex = 0;
            let maxCols = 0;
            data.slice(0, 5).forEach((row, idx) => {
                const cols = row.filter(c => c !== null && c !== undefined && String(c).trim() !== '').length;
                if (cols > maxCols) {
                    maxCols = cols;
                    headerRowIndex = idx;
                }
            });

            output += `### 推荐表头定义 (第 ${headerRowIndex + 1} 行)\n\n`;
            const headers = data[headerRowIndex] || [];
            output += `| 列索引 | 表头名称 | 第一行数据示例 |\n`;
            output += `|---|---|---|\n`;

            const sampleRow = data.length > headerRowIndex + 1 ? data[headerRowIndex + 1] : [];

            headers.forEach((h, i) => {
                if (h && String(h).trim() !== '') {
                    const sampleData = sampleRow[i] !== undefined ? String(sampleRow[i]).replace(/\n/g, ' ') : 'N/A';
                    output += `| ${i} | **${h}** | ${sampleData} |\n`;
                }
            });

            output += '\n### 观察与结论\n\n';
            output += '- 哪些列包含我们需要提取的 “商品名称”、“数量”、“单位”？\n';
            output += '- 哪一列可以作为 “周期提取” 的日期来源？\n';
            output += '- 前端解析时可能出现的陷阱是什么？\n';
        }
    });

    fs.writeFileSync(outputPath, output);
    console.log('分析完成，已保存在根目录下的 易派客明细结构分析.md');
} catch (error) {
    console.error('解析 Excel 失败:', error);
}
