import { Store } from 'pullstate'

type Props = {
  show: boolean,
  body: {}
}

const store = new Store<Props>({ 
  show: false, 
  body: {} 
})

export default store
