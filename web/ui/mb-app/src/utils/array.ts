export class ArrayUtils {
  groupBy<K, V>(array: V[], grouper: (item: V) => K) {
    return array.reduce((store, item) => {
      const key = grouper(item)
      if (!store.has(key)) {
        store.set(key, [item])
      } else {
        store.get(key)?.push(item)
      }
      return store
    }, new Map<K, V[]>())
  }

  groupMapBy<K, V>(list: V[], getKey: (item: V) => K) {
    const map = new Map<K, V[]>();
    list.forEach((item) => {
        const key = getKey(item);
        const collection = map.get(key);
        if (!collection) {
            map.set(key, [item]);
        } else {
            collection.push(item);
        }
    });
    return map;
  }
}

export default new ArrayUtils();