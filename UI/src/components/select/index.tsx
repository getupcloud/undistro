import React, { FC } from 'react'
import Select from 'react-select'

import './index.scss'

type OptionType = { value: any, label: string }

type Props = {
  label?: string,
  options: any,
  onChange: Function,
  value?: string
}

const SelectUndistro: FC<Props> = ({ 
  label,
  options,
  onChange,
  value
}) => {

  const handleChange = (option: any) => {
    onChange(option.value)
  }

  const getCorrectValue = ():OptionType => {
    return options.filter((elm: OptionType) => elm.value === value)[0]
  }

  return (
    <div className='select'>
    <div className='title-section'>
      <label>{label}</label>
    </div>

    <Select
      options={options}
      onChange={(e) => handleChange(e)}
      value={getCorrectValue()}
      classNamePrefix="select-container"
    />
  </div>
  )
}

export default SelectUndistro