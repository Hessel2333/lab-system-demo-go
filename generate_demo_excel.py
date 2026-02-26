import pandas as pd
from datetime import datetime
import os

def generate_demo_excel():
    # 模拟真实易派客导出的多列表头 (这里只列出包含关键信息的列，其余用空列填充，确保索引对齐)
    # 根据 ProcurementBatchImport.vue：
    # row[0] 订单编号
    # row[10] 下单时间
    # row[22] 物资类别
    # row[23] 商品名称
    # row[24] 商品类别
    # row[27] 采购数量
    # row[28] 基本计量单位
    
    # 易派客的表头通常是一二两行，为了简化，我们第一行留空或随便写，第二行写表头名，或者直接让第一行的名称也行
    columns = [f"Col_{i}" for i in range(30)]
    columns[0] = "订单编号"
    columns[10] = "下单时间"
    columns[22] = "物资类别"
    columns[23] = "商品名称"
    columns[24] = "商品类别"
    columns[27] = "采购数量"
    columns[28] = "基本计量单位"

    now_str = datetime.now().strftime("%Y-%m-%d %H:%M")
    
    # 构建数据行
    # 注意：我们的 "待办申购单" 里有: 甲醇, 硫酸, 超纯水
    data = [
        # 1. 完美匹配项 (CAS / 名称能匹配上)
        {0: "YPK-2026-0001", 10: now_str, 22: "化工", 23: "甲醇", 24: "试剂", 27: 5, 28: "瓶"},
        {0: "YPK-2026-0001", 10: now_str, 22: "化工", 23: "硫酸 (98%)", 24: "试剂", 27: 2, 28: "瓶"},
        {0: "YPK-2026-0001", 10: now_str, 22: "研发物资", 23: "超纯水", 24: "纯水试剂", 27: 10, 28: "L"},
        
        # 2. 杂项垃圾数据 (会被清洗掉，状态：已忽略)
        {0: "YPK-2026-0001", 10: now_str, 22: "劳保防护", 23: "丁腈乳胶手套(包)", 24: "耗材", 27: 20, 28: "包"},
        {0: "YPK-2026-0001", 10: now_str, 22: "办公", 23: "得力A4打印纸", 24: "文具", 27: 5, 28: "箱"},
        
        # 3. 紧急补录试剂 (系统未建档，需要人工指派认领人并补录)
        {0: "YPK-2026-0002", 10: now_str, 22: "化工", 23: "四氢呋喃(紧急未建档)", 24: "试剂", 27: 1, 28: "瓶"},
    ]

    # 将 dict 形式转为 df (填充空缺列为空字符串)
    df_data = []
    for row in data:
        row_arr = [''] * 30
        for k, v in row.items():
            row_arr[k] = v
        df_data.append(row_arr)

    # 易派客经常前几行是标题或者什么说明，所以我们手动插一行说明
    header_title = [''] * 30
    header_title[0] = "采购明细导出表"

    df = pd.DataFrame([header_title] + df_data, columns=columns)
    
    # 导出到项目根目录
    output_path = os.path.join(os.path.dirname(os.path.abspath(__file__)), "demo_procurement.xlsx")
    
    with pd.ExcelWriter(output_path, engine='openpyxl') as writer:
        df.to_excel(writer, index=False)
        print(f"Generated Excel file successfully at: {output_path}")

if __name__ == "__main__":
    generate_demo_excel()
