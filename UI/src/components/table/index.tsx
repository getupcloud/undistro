import React, { FC, useEffect, useState } from 'react'
import './index.scss'

type TypeHeader = {
  header: string[];
}

type TypeRow = {
  data: any,
  header: string[]
}

const data = ['coluna 1', 'coluna 2', 'coluna 3', 'coluna 4']
const header = ['header coluna 1', 'coluna 2', 'coluna 3', 'coluna 4']

const Table: FC = () => {
  return (
    <table className='table'>
      <thead>
        <HeaderColumn header={header} />
      </thead>
      
      <tbody>
        <Row data={data} header={header} />
      </tbody>
    </table>
  )
}

const HeaderColumn: FC<TypeHeader> = (header) => {
  return (
    <div className='header'>
      <label>{header}</label>
      <i className='arrow-up' />
    </div>
  )
}

const Row: FC<TypeRow> = ({ header, data }) => {
  const [rows, setRows] = useState<any>([])

  useEffect(() => {
    const headers = header.map((elm: any) => elm.header)
    setRows(Object.keys(data).filter(elm => headers.includes(elm)))
  }, [data, header])

  return (
    <tr>
      <td className='action-row'>...</td>
       {rows.map((key: any) => {
         return (
          <td key={key}>
            <span>{data[key]}</span>
          </td>
         )
       })}
    </tr>
  )
}


export default Table