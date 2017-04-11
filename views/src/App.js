import React, { Component } from 'react';
import {Router, Route, IndexRoute, browserHistory, Link} from 'react-router';
import request from 'superagent'
import './App.css';


class Main extends Component {
	render() {
		return (
      <div className="App">
				<LeftCol></LeftCol>
				<RightCol children={this.props.children}></RightCol>
      </div>
		);
	}
}

class LeftCol extends Component {
	render() {
		return (
			<div className="LeftCol">
				<Menu></Menu>
			</div>
		);
	}
}

class Menu extends Component {
	render() {
		return (
			<div className="Menu">
				<ul >
					<li><Link to="/projects">Projects</Link></li>
					<li><Link to="/register">Register Project</Link></li>
				</ul>
			</div>
		);
	}
}

class RightCol extends Component {
	render() {
		return (
			<div className="RightCol">
				{this.props.children}
			</div>
		);
	}
}

class Projects extends Component {
	constructor() {
		super();
    var url = "/list";
    request
      .get(url)
      .query({})
      .end(function(err, res){
        if (err) {
          alert(res.text);
					return;
        }
        var obj = JSON.parse(res.text);
        this.setState({projects: obj.projects});
      }.bind(this));
	}

	onClick(e, proj) {
    e.preventDefault();
    this.props.router.push({ pathname: 'detail/' + proj.name, state: { project: proj }});	
	}

	render() {
		var projects = [];
		
		if (this.state !== null) {
			projects = this.state.projects.map((proj, index) => {
				return (
					<ul key={index}>
						<li><Link to={"detail/" + proj.name} onClick={e => this.onClick(e, proj)}>{proj.name}</Link></li>
						<li>{proj.port}</li>
						<li>{proj.created_at}</li>
					</ul>
				);
			});
		}
		return (
			<div className="Projects">
					{projects}
			</div>
		);
	}
}

class Detail extends Component {
	onClickUnregisterProject(e, proj) {
		var obj = {
			name: proj.name
		}
    var url = "/unregister";
    request
      .post(url)
		  .type("form")
      .send({payload: JSON.stringify(obj)})
      .end(function(err, res){
        if (err) {
          alert(res.text);
					return;
        }
    		this.props.router.push({ pathname: 'projects' });	
      }.bind(this));
	}

	onClickUnregisterBranch(e, proj, branch) {
		var obj = {
			name: proj.name,
			branch: branch.name,
		}
    var url = "/down";
    request
      .post(url)
		  .type("form")
      .send({payload: JSON.stringify(obj)})
      .end(function(err, res){
        if (err) {
          alert(res.text);
					return;
        }
    		this.props.router.push({ pathname: 'projects' });	
      }.bind(this));
	}

	render() {
		var proj = this.props.location.state.project;
		var branch = [];
		if (proj.branch !== null) {
			branch = proj.branch.map((br, index) => {
				return (
					<li key={index}>
						<div>
							<ul>
								<li>{":" + br.port[0]}</li>
								<li>{br.name}</li>
								<li>{br.port.join(":")}</li>
								<li>{br.deployed_at}</li>
								<li><button onClick={e => this.onClickUnregisterBranch(e, proj, br)}>Down</button></li>
							</ul>
						</div>
					</li>
				);
			});
		}
		return (
			<div className="Detail">
				<ul>
					<li>Project Name: {proj.name}</li>
					<li>Repository Name: {proj.repo}</li>
					<li>Created at: {proj.created_at}</li>
					<li><button onClick={e => this.onClickUnregisterProject(e, proj)}>Unregister</button></li>
				</ul>
				Branch:<br/>
				<ul className="Branch">
					{branch}
				</ul>
			</div>
		);
	}
}

class Register extends Component {

	onChangeName(e) {
    this.setState({name: e.target.value});
	}

	onChangeRepoName(e) {
    this.setState({repo: e.target.value});
	}

	onSubmit(e) {
    var url = "/register";
    request
      .post(url)
		  .type("form")
      .send({payload: JSON.stringify(this.state)})
      .end(function(err, res){
        if (err) {
          alert(res.text);
					return;
        }
    		this.props.router.push({ pathname: 'projects' });	
      }.bind(this));
	}

	render() {
		return (
			<div className="Register">
				<ul>
					<li>Project Name: <input type="text" onChange={e => this.onChangeName(e)} /></li>
					<li>Repository Name: <input type="text" onChange={e => this.onChangeRepoName(e)} /></li>
				</ul>
			  <button onClick={e => this.onSubmit(e)}>Register</button>
			</div>
		);
	}
}

class App extends Component {
  render() {
    return (
      <Router history={browserHistory}>
        <Route path="/" component={Main}>
          <IndexRoute component={Projects}/ >
          <Route path="projects" component={Projects}/>
          <Route path="detail/:projectName" component={Detail}/>
          <Route path="register" component={Register}/>
        </Route>
      </Router>
    );
  }
}

export default App;
