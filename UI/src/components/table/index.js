import React, { useState, useEffect } from 'react'
import Classnames from 'classnames'
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
        {props.data.map((elm, i) => <Row delete={props.delete} onChange={props.onChange} header={props.header} key={i} data={elm} icon={props.icon} pause={props.pause} />)}
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
  const [show, setShow] = useState(false)

  useEffect(() => {
    const headers = props.header.map(elm => elm.field)
    setKeys(Object.keys(props.data).filter(elm => headers.includes(elm)))
  }, [props.data, props.header])

  return (
    <tr>
      <td className='select-row'>
        <i onClick={() => setShow(!show)} className='icon-dots' />
        {show && <ul>
          <li onClick={props.pause}>
            <i className={Classnames(props.icon ? 'icon-play' : 'icon-stop')} /> 
            {props.icon ? <p>Resume</p> : <p>Stop</p>}
          </li>
          <li><i className='icon-arrow-solid-up' /> <p>Update</p></li>
          <li><i className='icon-settings' /> <p>Settings</p></li>
          <li onClick={props.delete}><i className='icon-close-solid' /> <p>Delete</p></li>
			  </ul>}
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
