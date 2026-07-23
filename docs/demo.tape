Output docs/demo.gif
Output docs/demo.mp4
Set Shell "bash"
Set FontSize 18
Set Width 1400
Set Height 780
Set Theme "Catppuccin Mocha"
Set TypingSpeed 45ms

Hide
Type "source ./docs/demo-setup.sh || exit 1"
Enter
Type "clear"
Enter
Show

Type "# LeetMate — 不给答案，只给提示。"
Enter
Sleep 3s
Type "# Hints, not handouts. Practice now, review before you forget."
Enter
Sleep 3.5s
Hide
Type "clear && ./leetmate"
Enter
Show
Sleep 2.5s

# 首页（配音：终端里的刷题辅导工具）
Sleep 1.5s

# 选「开始刷题」→ 题单 → Hot 100（配音）
Down
Sleep 1.5s
Enter
Sleep 2.5s
Enter
Sleep 2.5s
Enter
# 进入 two-sum 题目（配音）
Sleep 3s

# 全屏看题目全文，突出读题（配音）
Type "o"
Sleep 5s
Escape
Sleep 1s

# 分级提示 · Hint：卡住时给方向，绝不泄答案（配音）
Type "1"
Sleep 7s

# 分级提示 · Nudge：再推一步，更具体但仍非完整代码（配音）
Type "2"
Sleep 7s

# 进内置编辑器，从头写代码（配音）
Type "i"
Sleep 1.5s
Set TypingSpeed 150ms
# seen = {}：方法体（2×Tab = 8 空格）
Enter
Tab
Tab
Type "seen = {}"
Sleep 1.5s
# for i, n in enumerate(nums):（2×Tab）
Enter
Tab
Tab
Type "for i, n in enumerate(nums):"
Sleep 1.5s
# if target - n in seen:：for 体内（3×Tab = 12 空格）
Enter
Tab
Tab
Tab
Type "if target - n in seen:"
Sleep 1.5s
# return [seen[target - n], i]：if 体内（4×Tab = 16 空格）
Enter
Tab
Tab
Tab
Tab
Type "return [seen[target - n], i]"
Sleep 1.5s
# seen[n] = i：for 体内（3×Tab）
Enter
Tab
Tab
Tab
Type "seen[n] = i"
Sleep 1.5s
# return []：方法体（2×Tab）
Enter
Tab
Tab
Type "return []"
Sleep 1.5s
Set TypingSpeed 45ms
# Esc 保存并退出编辑器
Escape
Sleep 2.5s

# 测试：远程跑用例，看 Accepted（配音）
Type "t"
Sleep 8s

# Review：请 Coach 审查代码，给反馈与建议（配音）
Type "3"
Sleep 6s

# 全屏展开辅导全文，突出审查（配音）
Type "o"
Sleep 1s
Tab
Sleep 8s

# 收起，提交到 LeetCode（配音）
Escape
Sleep 1.5s
Type "s"
Sleep 8s

# 收尾（配音总结）
Sleep 2.5s
Type "q"
Sleep 800ms

Hide
Type "cleanup_demo"
Enter
