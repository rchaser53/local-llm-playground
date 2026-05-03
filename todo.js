// todo.js

/**
 * Todoアプリケーションのコアロジック
 * タスクの配列と操作メソッドを定義します。
 */

let tasks = [];

/**
 * 新しいタスクを追加します。
 * @param {string} description - タスクの説明文
 */
function addTask(description) {
    if (!description || description.trim() === '') {
        console.log("エラー: タスクの説明を入力してください。");
        return;
    }
    const newTask = {
        id: Date.now(), // 簡単なID生成方法
        description: description.trim(),
        completed: false
    };
    tasks.push(newTask);
    console.log(`✅ タスクを追加しました: "${newTask.description}"`);
}

/**
 * すべてのタスクを表示します。
 * 完了したタスクは取り消し線（見た目上のシミュレーション）をつけます。
 */
function listTasks() {
    console.log("\n=====================");
    console.log("       TO DO LIST      ");
    console.log("=====================");

    if (tasks.length === 0) {
        console.log("登録されているタスクはありません。");
        console.log("=====================\n");
        return;
    }

    tasks.forEach((task, index) => {
        const status = task.completed ? '[DONE]' : '[TODO]';
        const displayDescription = task.completed ? `~~${task.description}~~` : task.description;
        console.log(`${index + 1}. ${status} ${displayDescription}`);
    });
    console.log("=====================\n");
}

/**
 * IDを指定してタスクの完了状態を切り替えます。
 * @param {number} taskId - 変更したいタスクのID
 */
function toggleTask(taskId) {
    const taskIndex = tasks.findIndex(task => task.id === taskId);
    if (taskIndex === -1) {
        console.log(`❌ エラー: ID ${taskId} のタスクは見つかりません。`);
        return;
    }

    tasks[taskIndex].completed = !tasks[taskIndex].completed;
    const status = tasks[taskIndex].completed ? "完了" : "未完了";
    console.log(`🔄 タスク ${taskIndex + 1} を ${status} に更新しました: "${tasks[taskIndex].description}"`);
}

/**
 * IDを指定してタスクを削除します。
 * @param {number} taskId - 削除したいタスクのID
 */
function deleteTask(taskId) {
    const initialLength = tasks.length;
    tasks = tasks.filter(task => task.id !== taskId);
    if (tasks.length < initialLength) {
        console.log(`🗑️ タスク ID ${taskId} を削除しました。`);
    } else {
        console.log(`❌ エラー: ID ${taskId} のタスクは見つかりませんでした。`);
    }
}


/**
 * コマンドラインインターフェース (CLI) ループをシミュレーションします。
 */
function runCLI() {
    console.log("\n🚀 Todo CLIが起動しました。以下のコマンドで操作してください:");
    console.log("  add [タスク内容]  : 新しいタスクを追加します。");
    console.log("  list              : 全てのタスクを表示します。");
    console.log("  toggle [タスクID] : タスクの状態（完了/未完了）を切り替えます。");
    console.log("  delete [タスクID] : タスクを削除します。");
    console.log("  exit              : アプリケーションを終了します。");

    const readline = require('readline');
    const rl = readline.createInterface({
        input: process.stdin,
        output: process.stdout
    });

    rl.on('line', (line) => {
        const parts = line.trim().split(/\s+/);
        const command = parts[0];
        const args = parts.slice(1);

        switch (command) {
            case 'add':
                // addコマンドの場合、残りの引数をすべて結合してタスク内容とする
                addTask(args.join(' '));
                break;
            case 'list':
                listTasks();
                break;
            case 'toggle':
                const toggleId = parseInt(args[0]);
                if (!isNaN(toggleId)) {
                    toggleTask(toggleId);
                } else {
                    console.log("エラー: 有効なタスクIDを入力してください。");
                }
                break;
            case 'delete':
                const deleteId = parseInt(args[0]);
                if (!isNaN(deleteId)) {
                    deleteTask(deleteId);
                } else {
                    console.log("エラー: 有効なタスクIDを入力してください。");
                }
                break;
            case 'exit':
                rl.close();
                break;
            default:
                console.log(`不明なコマンドです: ${command}`);
                break;
        }
    });

    rl.on('close', () => {
        console.log("🚀 Todo CLIを終了しました。バイバイ！");
        process.exit(0);
    });
}

// アプリケーションのエントリポイント
runCLI();
