export interface ReadHistoryItem {
    seriesId: string
    seriesSlug: string
    seriesTitle: string
    chapterId: string
    chapterSlug: string
    chapterTitle: string
    readAt: number
}

const STORAGE_KEY = 'mhentai_read_history'
const MAX_HISTORY_ITEMS = 50

function parseHistory(): ReadHistoryItem[] {
    try {
        const raw = localStorage.getItem(STORAGE_KEY)
        if (!raw) return []
        const items = JSON.parse(raw) as ReadHistoryItem[]
        if (!Array.isArray(items)) return []
        return items
    } catch {
        return []
    }
}

function saveHistory(items: ReadHistoryItem[]) {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(items.slice(0, MAX_HISTORY_ITEMS)))
}

export function getReadHistory(): ReadHistoryItem[] {
    return parseHistory()
}

export function getReadHistoryForSeries(seriesId: string, limit = 3): ReadHistoryItem[] {
    return parseHistory()
        .filter(item => item.seriesId === seriesId)
        .sort((a, b) => b.readAt - a.readAt)
        .slice(0, limit)
}

export function addReadHistory(entry: ReadHistoryItem) {
    const history = parseHistory()
    const existingIndex = history.findIndex(item => item.chapterId === entry.chapterId)
    if (existingIndex !== -1) {
        history.splice(existingIndex, 1)
    }
    history.unshift(entry)
    saveHistory(history)
}

export function isChapterRead(chapterSlug: string): boolean {
    return parseHistory().some(item => item.chapterSlug === chapterSlug)
}
