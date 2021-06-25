import React, { useState, useEffect } from 'react'
import './index.scss'

const Table = (props) => {
  return (
    <table className='table'>
      <thead>
        <tr>
          <td />
          {props.header.map(elm => <td key={elm.name}><ColumnHeader data={elm} /></td>)}
        </tr>
      </thead>
      <tbody>
        {props.data.map((elm, i) => <Row onChange={props.onChange} header={props.header} key={i} data={elm} onClick={props.onClick} />)}
      </tbody>
    </table>
  )
}

const ColumnHeader = (props) => {
  const [order, setOrder] = useState(true)

  const handleChangeOrder = (value) => {
    setOrder(value)
  }

  return (
    <div className='header'>
      <label>{props.data.name}</label>
      <i onClick={() => handleChangeOrder(!order)} className='' />
    </div>
  )
}

const Row = (props) => {
  const [keys, setKeys] = useState([])

  useEffect(() => {
    const headers = props.header.map(elm => elm.field)
    setKeys(Object.keys(props.data).filter(elm => headers.includes(elm)))
  }, [props.data, props.header])

  return (
    <tr onClick={props.onClick}>
      <td className='select-row'>
        <i className='icon-dots' />
      </td>
      {keys.map((key) => {
        return (
          <td key={key}>
            <div>
              <span>{props.data[key]}</span>
            </div>
          </td>
        )
      })}
    </tr>
  )
}

export default Table
