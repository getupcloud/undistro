import store from '../modals/store'

class Modals {
  default (body: {}) {
    this.show('test', body)
  }

  show (id: string, body: {}) {
    store.update((s: any) => {
      s.id = id
      s.show = true
      s.body = body
    })
  }

  hide () {
    store.update((s: any) => {
      s.id = null
      s.show = false
      s.body = {}
    })
  }
}

export default new Modals()