import React, {Component} from 'react'
import axios from "axios"
import { Header, Form, Input } from 'semantic-ui-react'

let endpoint = 'http://localhost:8080'

class Stats extends Component {
  constructor(props) {
    super(props)

    this.state = {
      mini: '',
      res: '',
      hits: '',
      submitted: false,
    }
  }

  componentDidMount() {
    
  }

  onChange = event => {
    this.setState({
      [event.target.name]: event.target.value
    })
  }

  onSubmit = () => {
    let {mini} = this.state 
    mini = mini.substring(mini.length-6, mini.length)
    
    if (mini) {
      axios.post(
        endpoint+'/api/current',
        {mini},
        {
          headers: {
            'Content-Type': 'application/json'
          }
        }
      ).then(res => {
        this.setState({res: res, submitted: true})
      }).catch(e => {
        if(e.response) {
          this.setState({res: e.response, submitted: true})
        }
        else {
          this.setState({submitted: true})
        }
      })
    }
  }

  isMini = () => {
    let res = this.state.res
    if(!this.state.submitted){
      return
    }
    if (res.status === 200) {
      console.log(res.data)
      return (
        <div>
          <p><strong>MiniURL: </strong><a href={endpoint+'/'+res.data.mini}>{endpoint+'/'+res.data.mini}</a></p>
          <p><strong>Original URL: </strong><a href={res.data.long}>{res.data.long}</a></p>
          <p><strong>Hits: </strong>{res.data.Hits}</p>
        </div>
        
      )
    }else if (res.status === 404) {
      return (
        <div className='ui negative message'>
          <i className='close icon'></i>
          <div className='header'>
            Could not find that MiniURL
          </div>
        </div>
      )
    } else if(!this.state.submitted){
      return
    } else {
      return (
        <div className='ui negative message'>
          <i className='close icon'></i>
          <div className='header'>
            An error occured
          </div>
        </div>
      )
    }
  }

  render() {
    const styles = {
      padding: '2em',
    }

    return (
      <div>
        <div className='row' style={styles}>
          <Header className='header' as='h1'>
            MiniURL Statistics
          </Header>
        </div>
        <div className='row' style={styles}>
          <Form onSubmit={this.onSubmit}>
            <Input type='text' name='mini' value={this.state.mini} onChange={this.onChange} fluid placeholder='Enter miniURL' />
          </Form>
        </div>
        <button onClick={this.onSubmit}>Get Stats</button>
        {this.isMini()}
      </div>
    )
  }
}

export default Stats;